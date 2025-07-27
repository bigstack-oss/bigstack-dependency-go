package openstack

import (
	"context"
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/domains"
)

func (h *Helper) IsDomainExists(name string) (bool, error) {
	domains, err := h.ListDomains(&domains.ListOpts{Name: name})
	if err != nil {
		return false, err
	}

	for _, domain := range domains {
		if domain.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func (h *Helper) ListDomains(opts *domains.ListOpts) ([]domains.Domain, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := domains.List(h.Identity, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	domains, err := domains.ExtractDomains(pages)
	if err != nil {
		return nil, err
	}

	if len(domains) == 0 {
		return nil, fmt.Errorf(
			"no domains found with the provided options: %v", opts,
		)
	}

	return domains, nil
}
