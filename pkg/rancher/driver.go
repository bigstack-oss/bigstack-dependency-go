package rancher

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
)

type NodeDriversResp struct {
	NodeDrivers []NodeDriver `json:"data"`
}

type NodeDriver struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

func (h *Helper) ActivateNodeDriver(name string) error {
	u, err := url.Parse(h.Options.Url)
	if err != nil {
		return err
	}

	u.Path = fmt.Sprintf("/v3/nodeDrivers/%s", name)
	u.RawQuery = "action=activate"
	resp, err := h.Http.R().
		SetHeaders(GenAuthHeaders(h.Options.Token)).
		Post(u.String())
	if err != nil {
		return err
	}

	if !resp.IsError() {
		return nil
	}

	return fmt.Errorf(
		"failed to activate node driver %s (%s %s)",
		name,
		u.String(),
		resp.String(),
	)
}

func (h *Helper) WaitNodeDriverStatus(name, state string, timeout int) error {
	u, err := url.Parse(h.Options.Url)
	if err != nil {
		return err
	}

	u.Path = "/v3/nodedrivers"
	for range timeout {
		wait.Seconds(1)

		listResp := &NodeDriversResp{}
		resp, err := h.Http.R().
			SetResult(listResp).
			SetHeaders(GenAuthHeaders(h.Options.Token)).
			Get(u.String())
		if err != nil {
			continue
		}

		if resp.IsError() {
			continue
		}

		if h.isNodeDriverMatchedStatus(listResp.NodeDrivers, name, state) {
			return nil
		}
	}

	return fmt.Errorf(
		"failed to wait node driver %s status %s",
		name,
		state,
	)
}

func (h *Helper) isNodeDriverMatchedStatus(drivers []NodeDriver, name, state string) bool {
	for _, driver := range drivers {
		if driver.Name != name {
			continue
		}

		if strings.EqualFold(driver.State, state) {
			return true
		}
	}

	return false
}
