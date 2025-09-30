package openstack

import (
	"context"
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/hypervisors"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/quotasets"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
)

func (h *Helper) UpdateComputeQuotas(projectId string, opts quotasets.UpdateOpts) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return quotasets.Update(ctx, h.Compute, projectId, opts).Err
}

func (h *Helper) ListServers(opts servers.ListOpts) ([]servers.Server, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := servers.List(h.Compute, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	return servers.ExtractServers(pages)
}

func (h *Helper) GetServer(id string) (*servers.Server, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return servers.Get(ctx, h.Compute, id).Extract()
}

func (h *Helper) DeleteServer(id string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return servers.Delete(ctx, h.Compute, id).Err
}

func (h *Helper) GetHypervisorStatistics() (*hypervisors.Statistics, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return hypervisors.GetStatistics(ctx, h.Compute).Extract()
}

func (h *Helper) ListHypervisors(opts hypervisors.ListOpts) ([]hypervisors.Hypervisor, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := hypervisors.List(h.Compute, hypervisors.ListOpts{}).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	return hypervisors.ExtractHypervisors(pages)
}

func (h *Helper) GetHypervisorByHostname(hostname string) (*hypervisors.Hypervisor, error) {
	hps, err := h.ListHypervisors(hypervisors.ListOpts{})
	if err != nil {
		return nil, err
	}

	err = fmt.Errorf("hypervisor with hostname %s not found", hostname)
	if len(hps) == 0 {
		return nil, err
	}

	for _, hypervisor := range hps {
		if hypervisor.HypervisorHostname == hostname {
			return &hypervisor, nil
		}
	}

	return nil, err
}

func (h *Helper) GetHypervisorUpTime(id string) (*hypervisors.Uptime, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return hypervisors.GetUptime(ctx, h.Compute, id).Extract()
}

func (h *Helper) IsImageExistByName(name string) (bool, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	opts := images.ListOpts{Name: name}
	pages, err := images.List(h.Image, opts).AllPages(ctx)
	if err != nil {
		return false, err
	}

	list, err := images.ExtractImages(pages)
	if err != nil {
		return false, err
	}

	for _, image := range list {
		if image.Name == name {
			return true, nil
		}
	}

	return false, fmt.Errorf(
		"image %s not found",
		name,
	)
}

func (h *Helper) IsImageExist(id string) (bool, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	opts := images.ListOpts{ID: id}
	pages, err := images.List(h.Image, opts).AllPages(ctx)
	if err != nil {
		return false, err
	}

	list, err := images.ExtractImages(pages)
	if err != nil {
		return false, err
	}

	for _, image := range list {
		if image.ID == id {
			return true, nil
		}
	}

	return false, fmt.Errorf(
		"image %s not found",
		id,
	)
}

func (h *Helper) ListImages(opts images.ListOpts) ([]images.Image, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := images.List(h.Image, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	return images.ExtractImages(pages)
}

func (h *Helper) GetImage(id string) (*images.Image, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return images.Get(ctx, h.Image, id).Extract()
}

func (h *Helper) GetImageByName(name string) (*images.Image, error) {
	images, err := h.ListImages(images.ListOpts{Name: name})
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("no images found")
	}

	for _, image := range images {
		if image.Name == name {
			return &image, nil
		}
	}

	return nil, fmt.Errorf(
		"image with name %s not found", name,
	)
}

func (h *Helper) UpdateImageProperty(id string, opts images.UpdateOpts) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return images.Update(ctx, h.Image, id, opts).Err
}

func (h *Helper) IsFlavorExist(name string) (bool, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	pages, err := flavors.ListDetail(h.Compute, flavors.ListOpts{}).AllPages(ctx)
	if err != nil {
		return false, err
	}

	list, err := flavors.ExtractFlavors(pages)
	if err != nil {
		return false, err
	}

	for _, flavor := range list {
		if flavor.Name == name {
			return true, nil
		}
	}

	return false, fmt.Errorf(
		"flavor %s not found",
		name,
	)
}

func (h *Helper) UpdateServer(id string, opts servers.UpdateOpts) (*servers.Server, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	res := servers.Update(ctx, h.Compute, id, opts)
	if res.Err != nil {
		return nil, res.Err
	}
	return res.Extract()
}
