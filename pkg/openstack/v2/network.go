package openstack

import (
	"context"
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/dns/v2/recordsets"
	"github.com/gophercloud/gophercloud/v2/openstack/dns/v2/zones"
	"github.com/gophercloud/gophercloud/v2/openstack/loadbalancer/v2/flavorprofiles"
	"github.com/gophercloud/gophercloud/v2/openstack/loadbalancer/v2/loadbalancers"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/quotas"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/security/rules"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/sharenetworks"
)

func (h *Helper) ListNetworks(opts networks.ListOpts) ([]networks.Network, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := networks.List(h.Network, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	return networks.ExtractNetworks(pages)
}

func (h *Helper) DeleteNetwork(id string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return networks.Delete(ctx, h.Network, id).Err
}

func (h *Helper) GetNetworkByName(opts networks.ListOpts) (*networks.Network, error) {
	networks, err := h.ListNetworks(opts)
	if err != nil {
		return nil, err
	}

	for _, network := range networks {
		if network.Name == opts.Name {
			return &network, nil

		}
	}

	return nil, fmt.Errorf("network %s not found", opts.Name)
}

func (h *Helper) CreateNetwork(opts networks.CreateOpts) (*networks.Network, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return networks.Create(ctx, h.Network, opts).Extract()
}

func (h *Helper) ListSubnets(opts subnets.ListOpts) ([]subnets.Subnet, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := subnets.List(h.Network, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	return subnets.ExtractSubnets(pages)
}

func (h *Helper) CreateSubnet(opts subnets.CreateOpts) (*subnets.Subnet, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return subnets.Create(ctx, h.Network, opts).Extract()
}

func (h *Helper) DeleteSubnet(id string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return subnets.Delete(ctx, h.Network, id).Err
}

func (h *Helper) ListRouters(opts routers.ListOpts) ([]routers.Router, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := routers.List(h.Network, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	return routers.ExtractRouters(pages)
}

func (h *Helper) CreateRouter(opts routers.CreateOpts) (*routers.Router, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return routers.Create(ctx, h.Network, opts).Extract()
}

func (h *Helper) UpdateRouter(id string, opts routers.UpdateOpts) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return routers.Update(ctx, h.Network, id, opts).Err
}

func (h *Helper) AttachNetworkToRouter(id string, opts routers.AddInterfaceOpts) (*routers.InterfaceInfo, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return routers.AddInterface(ctx, h.Network, id, opts).Extract()
}

func (h *Helper) DeleteRouter(id string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return routers.Delete(ctx, h.Network, id).Err
}

func (h *Helper) DeleteRouterInterface(routerId string, opts routers.RemoveInterfaceOpts) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return routers.RemoveInterface(
		ctx,
		h.Network,
		routerId,
		opts,
	).Err
}

func (h *Helper) ListSecurityGroups(opts groups.ListOpts) ([]groups.SecGroup, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := groups.List(h.Network, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	return groups.ExtractGroups(pages)
}

func (h *Helper) CreateSecurityGroup(opts groups.CreateOpts) (*groups.SecGroup, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return groups.Create(ctx, h.Network, opts).Extract()
}

func (h *Helper) CreateSecurityGroupRule(opts rules.CreateOpts) (*rules.SecGroupRule, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return rules.Create(ctx, h.Network, opts).Extract()
}

func (h *Helper) UpdateNetworkQuotas(projectId string, opts quotas.UpdateOpts) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return quotas.Update(ctx, h.Network, projectId, opts).Err
}

func (h *Helper) GetPortByIp(ip string) (*ports.Port, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := ports.List(
		h.Network,
		ports.ListOpts{FixedIPs: []ports.FixedIPOpts{{IPAddress: ip}}},
	).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	ports, err := ports.ExtractPorts(pages)
	if err != nil {
		return nil, err
	}

	for _, port := range ports {
		for _, fixedIP := range port.FixedIPs {
			if fixedIP.IPAddress == ip {
				return &port, nil
			}
		}
	}

	return nil, fmt.Errorf("port with ip %s not found", ip)
}

func (h *Helper) ListPorts(opts ports.ListOpts) ([]ports.Port, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := ports.List(h.Network, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	return ports.ExtractPorts(pages)
}

func (h *Helper) DeletePort(id string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return ports.Delete(ctx, h.Network, id).Err
}

func (h *Helper) GetSubnetByName(opts subnets.ListOpts) (*subnets.Subnet, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := subnets.List(h.Network, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	subnets, err := subnets.ExtractSubnets(pages)
	if err != nil {
		return nil, err
	}

	for _, subnet := range subnets {
		if subnet.Name == opts.Name {
			return &subnet, nil
		}
	}

	return nil, fmt.Errorf("subnet %s not found", opts.Name)
}

func (h *Helper) GetSecurityGroupByName(opts groups.ListOpts) (*groups.SecGroup, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := groups.List(h.Network, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	secGroups, err := groups.ExtractGroups(pages)
	if err != nil {
		return nil, err
	}

	for _, secGroup := range secGroups {
		if secGroup.Name == opts.Name {
			return &secGroup, nil
		}
	}

	return nil, fmt.Errorf("security group %s not found", opts.Name)
}

func (h *Helper) GetSecurityGroup(id string) (*groups.SecGroup, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return groups.Get(ctx, h.Network, id).Extract()
}

func (h *Helper) DeleteSecurityGroupRule(id string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return rules.Delete(ctx, h.Network, id).Err
}

func (h *Helper) GetShareNetworkByName(opts sharenetworks.ListOpts) (*sharenetworks.ShareNetwork, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := sharenetworks.ListDetail(h.Share, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	shareNetworks, err := sharenetworks.ExtractShareNetworks(pages)
	if err != nil {
		return nil, err
	}

	for _, shareNetwork := range shareNetworks {
		if shareNetwork.Name == opts.Name {
			return &shareNetwork, nil
		}
	}

	return nil, fmt.Errorf("share network %s not found", opts.Name)
}

func (h *Helper) CreateShareNetwork(client *gophercloud.ServiceClient, opts sharenetworks.CreateOpts) (*sharenetworks.ShareNetwork, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return sharenetworks.Create(ctx, client, opts).Extract()
}

func (h *Helper) DeleteShareNetwork(client *gophercloud.ServiceClient, id string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return sharenetworks.Delete(ctx, client, id).Err
}

func (h *Helper) ListLoadBalancers(opts loadbalancers.ListOpts) ([]loadbalancers.LoadBalancer, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := loadbalancers.List(h.Loadbalancer, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	return loadbalancers.ExtractLoadBalancers(pages)
}

func (h *Helper) GetLoadBalancer(id string) (*loadbalancers.LoadBalancer, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return loadbalancers.Get(ctx, h.Loadbalancer, id).Extract()
}

func (h *Helper) ListFloatingIps(opts floatingips.ListOpts) ([]floatingips.FloatingIP, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := floatingips.List(h.Network, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	return floatingips.ExtractFloatingIPs(pages)
}

func (h *Helper) DisassociateFloatingIp(id string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return floatingips.Update(
		ctx,
		h.Network,
		id,
		floatingips.UpdateOpts{PortID: nil},
	).Err
}

func (h *Helper) DeleteLoadBalancer(id string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return loadbalancers.Delete(
		ctx,
		h.Loadbalancer,
		id,
		loadbalancers.DeleteOpts{Cascade: true},
	).Err
}

func (h *Helper) ListFlavorProfiles(opts flavorprofiles.ListOpts) ([]flavorprofiles.FlavorProfile, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := flavorprofiles.List(h.Loadbalancer, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	return flavorprofiles.ExtractFlavorProfiles(pages)
}

func (h *Helper) DeleteFloatingIP(id string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return floatingips.Delete(ctx, h.Network, id).Err
}

func (h *Helper) CreateDnsZone(opts zones.CreateOpts) (*zones.Zone, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return zones.Create(ctx, h.Dns, opts).Extract()
}

func (h *Helper) DeleteDnsZone(id string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return zones.Delete(ctx, h.Dns, id).Err
}

func (h *Helper) ListDnsZones(opts zones.ListOpts) ([]zones.Zone, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := zones.List(h.Dns, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	return zones.ExtractZones(pages)
}

func (h *Helper) GetDnsZoneByName(name string) (*zones.Zone, error) {
	zones, err := h.ListDnsZones(zones.ListOpts{Name: name})
	if err != nil {
		return nil, err
	}

	for _, zone := range zones {
		if zone.Name == name {
			return &zone, nil
		}
	}

	return nil, fmt.Errorf(
		"dns zone %s not found",
		name,
	)
}

func (h *Helper) CreateDnsRecord(zoneId string, opts recordsets.CreateOpts) (*recordsets.RecordSet, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return recordsets.Create(ctx, h.Dns, zoneId, opts).Extract()
}

func (h *Helper) DeleteDnsRecord(zoneId, id string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return recordsets.Delete(ctx, h.Dns, zoneId, id).Err
}
