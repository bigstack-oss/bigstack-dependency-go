package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
)

// Ec2Helper wraps an EC2 client scoped to one (credentials, region). AWS
// resources are region-scoped and CMP sources per-project credentials, so
// callers build one Ec2Helper per (project credentials, region) rather than
// reusing the global S3 helper.
type Ec2Helper struct {
	Client *ec2.Client
}

// NewEc2Helper builds an EC2 client from static credentials and a region.
func NewEc2Helper(accessKey, secretKey, region string) (*Ec2Helper, error) {
	cfg, err := newAwsConfig(Options{
		Region:            region,
		AccessKey:         accessKey,
		SecretKey:         secretKey,
		EnableStaticCreds: true,
	})
	if err != nil {
		return nil, err
	}

	return &Ec2Helper{Client: ec2.NewFromConfig(*cfg)}, nil
}

func (h *Ec2Helper) TerminateInstance(id string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	_, err := h.Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{id},
	})
	return err
}

func (h *Ec2Helper) WaitInstanceTerminated(id string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(300))
	defer cancel()

	waiter := ec2.NewInstanceTerminatedWaiter(h.Client)
	return waiter.Wait(
		ctx,
		&ec2.DescribeInstancesInput{InstanceIds: []string{id}},
		5*time.Minute,
	)
}

func (h *Ec2Helper) DeleteVolume(id string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	_, err := h.Client.DeleteVolume(ctx, &ec2.DeleteVolumeInput{VolumeId: &id})
	return err
}

func (h *Ec2Helper) ReleaseAddress(allocationID string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	_, err := h.Client.ReleaseAddress(ctx, &ec2.ReleaseAddressInput{
		AllocationId: &allocationID,
	})
	return err
}

func (h *Ec2Helper) DeleteSecurityGroup(id string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	_, err := h.Client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
		GroupId: &id,
	})
	return err
}

// DeleteVpc removes a VPC after tearing down the dependencies AWS won't cascade
// — internet gateways, non-main route-table associations, non-default network
// ACLs, and subnets (with their network interfaces) — otherwise DeleteVpc fails
// with a DependencyViolation. Mirrors the worker's delete_vpc.
func (h *Ec2Helper) DeleteVpc(vpcID string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(120))
	defer cancel()

	if err := h.detachInternetGateways(ctx, vpcID); err != nil {
		return err
	}
	if err := h.disassociateRouteTables(ctx, vpcID); err != nil {
		return err
	}
	if err := h.deleteNetworkACLs(ctx, vpcID); err != nil {
		return err
	}
	if err := h.deleteSubnets(ctx, vpcID); err != nil {
		return err
	}

	_, err := h.Client.DeleteVpc(ctx, &ec2.DeleteVpcInput{VpcId: &vpcID})
	return err
}

func vpcFilter(vpcID string) types.Filter {
	return types.Filter{Name: aws.String("vpc-id"), Values: []string{vpcID}}
}

func (h *Ec2Helper) detachInternetGateways(ctx context.Context, vpcID string) error {
	out, err := h.Client.DescribeInternetGateways(ctx, &ec2.DescribeInternetGatewaysInput{
		Filters: []types.Filter{{Name: aws.String("attachment.vpc-id"), Values: []string{vpcID}}},
	})
	if err != nil {
		return err
	}

	for _, igw := range out.InternetGateways {
		if _, err := h.Client.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{
			InternetGatewayId: igw.InternetGatewayId,
			VpcId:             &vpcID,
		}); err != nil {
			return err
		}
		if _, err := h.Client.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{
			InternetGatewayId: igw.InternetGatewayId,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (h *Ec2Helper) disassociateRouteTables(ctx context.Context, vpcID string) error {
	out, err := h.Client.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{
		Filters: []types.Filter{vpcFilter(vpcID)},
	})
	if err != nil {
		return err
	}

	for _, rt := range out.RouteTables {
		for _, assoc := range rt.Associations {
			if assoc.Main != nil && *assoc.Main {
				continue
			}
			if _, err := h.Client.DisassociateRouteTable(ctx, &ec2.DisassociateRouteTableInput{
				AssociationId: assoc.RouteTableAssociationId,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *Ec2Helper) deleteNetworkACLs(ctx context.Context, vpcID string) error {
	out, err := h.Client.DescribeNetworkAcls(ctx, &ec2.DescribeNetworkAclsInput{
		Filters: []types.Filter{vpcFilter(vpcID)},
	})
	if err != nil {
		return err
	}

	for _, acl := range out.NetworkAcls {
		if acl.IsDefault != nil && *acl.IsDefault {
			continue
		}
		if _, err := h.Client.DeleteNetworkAcl(ctx, &ec2.DeleteNetworkAclInput{
			NetworkAclId: acl.NetworkAclId,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (h *Ec2Helper) deleteSubnets(ctx context.Context, vpcID string) error {
	out, err := h.Client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []types.Filter{vpcFilter(vpcID)},
	})
	if err != nil {
		return err
	}

	for _, subnet := range out.Subnets {
		enis, err := h.Client.DescribeNetworkInterfaces(ctx, &ec2.DescribeNetworkInterfacesInput{
			Filters: []types.Filter{{Name: aws.String("subnet-id"), Values: []string{*subnet.SubnetId}}},
		})
		if err != nil {
			return err
		}
		for _, eni := range enis.NetworkInterfaces {
			if _, err := h.Client.DeleteNetworkInterface(ctx, &ec2.DeleteNetworkInterfaceInput{
				NetworkInterfaceId: eni.NetworkInterfaceId,
			}); err != nil {
				return err
			}
		}
		if _, err := h.Client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{SubnetId: subnet.SubnetId}); err != nil {
			return err
		}
	}
	return nil
}
