package openstack

import (
	"context"
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/gophercloud/gophercloud/v2"
	blockQuotas "github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/quotasets"
	computeQuotas "github.com/gophercloud/gophercloud/v2/openstack/compute/v2/quotasets"
	netQuotas "github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/quotas"
)

type ManilaQuotaSet struct {
	Shares            *int `json:"shares,omitempty"`
	Gigabytes         *int `json:"gigabytes,omitempty"`
	Snapshots         *int `json:"snapshots,omitempty"`
	SnapshotGigabytes *int `json:"snapshot_gigabytes,omitempty"`
	ShareNetworks     *int `json:"share_networks,omitempty"`
}

func (h *Helper) UpdateComputeQuota(projectID string, updateOpts computeQuotas.UpdateOptsBuilder) (*computeQuotas.QuotaSet, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return computeQuotas.Update(ctx, h.Compute, projectID, updateOpts).Extract()
}

func (h *Helper) UpdateBlockStorageQuota(projectID string, updateOpts blockQuotas.UpdateOptsBuilder) (*blockQuotas.QuotaSet, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return blockQuotas.Update(ctx, h.Storage, projectID, updateOpts).Extract()
}

func (h *Helper) UpdateNetworkQuota(projectID string, updateOpts netQuotas.UpdateOptsBuilder) (*netQuotas.Quota, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return netQuotas.Update(ctx, h.Network, projectID, updateOpts).Extract()
}

func (h *Helper) UpdateSharedFileSystemQuota(projectID string, updateOpts ManilaQuotaSet) error {
	type requestBody struct {
		QuotaSet ManilaQuotaSet `json:"quota_set"`
	}
	body := requestBody{
		QuotaSet: updateOpts,
	}

	url := h.Share.ServiceURL("os-quota-sets", projectID)
	resp, err := h.Share.Put(context.TODO(), url, body, nil, &gophercloud.RequestOpts{
		OkCodes: []int{200, 201, 202},
	})
	if err != nil {
		return fmt.Errorf("failed to update manila quotas: %w", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
