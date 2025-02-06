package openstack

import (
	"context"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/quotasets"
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shares"
)

func (h *Helper) UpdateStorageQuotas(projectId string, opts quotasets.UpdateOpts) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return quotasets.Update(ctx, h.Storage, projectId, opts).Err
}

func (h *Helper) ListVolumes(opts volumes.ListOpts) ([]volumes.Volume, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := volumes.List(h.Storage, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	return volumes.ExtractVolumes(pages)
}

func (h *Helper) GetVolume(volumeId string) (*volumes.Volume, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return volumes.Get(ctx, h.Storage, volumeId).Extract()
}

func (h *Helper) ListShares(opts shares.ListOpts) ([]shares.Share, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := shares.ListDetail(h.Share, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	return shares.ExtractShares(pages)
}
