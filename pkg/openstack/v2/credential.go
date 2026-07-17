package openstack

import (
	"context"
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/credentials"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/tokens"
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

func (h *Helper) ListEc2Credentials(opts credentials.ListOpts) ([]credentials.Credential, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := credentials.List(h.Identity, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	return credentials.ExtractCredentials(pages)
}

func (h *Helper) DeleteEc2Credential(id string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return credentials.Delete(ctx, h.Identity, id).ExtractErr()
}

func (h *Helper) ValidToken(token string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return tokens.Get(ctx, h.Identity, token).Err
}
