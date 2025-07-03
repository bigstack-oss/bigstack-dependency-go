package rancher

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type CloudCredential struct {
	Id                        string `json:"id"`
	Uuid                      string `json:"uuid"`
	Name                      string `json:"name"`
	Type                      string `json:"type"`
	BaseType                  string `json:"baseType"`
	Metadata                  `json:"metadata"`
	Annotations               `json:"annotations"`
	OpenstackCredentialConfig `json:"openstackcredentialConfig"`
	Created                   string `json:"created"`
}

type Annotations struct {
	ProvisioningCattleIoDriver string `json:"provisioning.cattle.io/driver,omitempty"`
}

type OpenstackCredentialConfig struct {
	Password string `json:"password"`
}

func (c *CloudCredential) Bytes() ([]byte, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (h *Helper) CreateCloudCredential(credential *CloudCredential) (*CloudCredential, error) {
	u, err := url.Parse(h.Options.Url)
	if err != nil {
		return nil, err
	}

	u.Path = "/v3/cloudcredentials"
	b, err := credential.Bytes()
	if err != nil {
		return nil, err
	}

	cloudCredential := &CloudCredential{}
	resp, err := h.Http.R().
		SetResult(cloudCredential).
		SetHeaders(GenAuthHeaders(h.Options.Token)).
		SetBody(string(b)).
		Post(u.String())
	if err != nil {
		return nil, err
	}

	if !resp.IsError() {
		return cloudCredential, nil
	}

	return nil, fmt.Errorf(
		"failed to create cloud credential(%d %s)",
		resp.StatusCode(),
		resp.String(),
	)
}
