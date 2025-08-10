package rancher

import (
	"fmt"
	"net/url"
)

func (h *Helper) ActiveNodeDriver(name string) error {
	u, err := url.Parse(h.Options.Url)
	if err != nil {
		return err
	}

	u.Path = fmt.Sprintf("/v3/nodeDrivers/%s?action=activate", name)
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
		"failed to activate node driver %s(%s)",
		name,
		resp.String(),
	)
}
