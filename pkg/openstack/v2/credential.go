package openstack

import (
	"context"
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/credentials"
)

func (h *Helper) CreateEc2Credential(userId, projectId, accessKey, secretKey string) (*credentials.Credential, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return credentials.Create(
		ctx,
		h.Identity,
		credentials.CreateOpts{
			UserID:    userId,
			ProjectID: projectId,
			Type:      "ec2",
			Blob:      fmt.Sprintf("{\"access\":\"%s\",\"secret\":\"%s\"}", accessKey, secretKey),
		},
	).Extract()
}
