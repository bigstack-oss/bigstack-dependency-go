package openstack

import (
	"context"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/tokens"
)

func (h *Helper) GetServiceCatalog() (*tokens.ServiceCatalog, error) {
	id, err := h.Provider.GetAuthResult().ExtractTokenID()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return tokens.Get(ctx, h.Identity, id).ExtractServiceCatalog()
}
