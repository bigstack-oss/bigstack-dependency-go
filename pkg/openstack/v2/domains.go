package openstack

import (
	"context"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/domains"
)

func (h *Helper) ListDomains(opts *domains.ListOpts) ([]domains.Domain, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := domains.List(h.Identity, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	return domains.ExtractDomains(pages)
}
