package rancher

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type OpenstackMachine struct {
	ActiveTimeout  string `json:"activeTimeout"`
	AuthUrl        string `json:"authUrl"`
	EndpointType   string `json:"endpointType"`
	FlavorName     string `json:"flavorName"`
	FloatingipPool string `json:"floatingipPool"`
	ImageName      string `json:"imageName"`
	Insecure       bool   `json:"insecure"`
	IpVersion      string `json:"ipVersion"`
	Metadata       `json:"metadata"`
	NetId          string `json:"netId"`
	SecGroups      string `json:"secGroups"`
	SshPort        string `json:"sshPort"`
	SshUser        string `json:"sshUser"`
	TenantId       string `json:"tenantId"`
	UserId         string `json:"userId"`
	UserDataFile   string `json:"userDataFile,omitempty"`
	Type           string `json:"type"`
}

type Metadata struct {
	Name            string `json:"name"`
	Annotations     `json:"annotations"`
	Finalizers      []string          `json:"finalizers,omitempty"`
	GenerateName    string            `json:"generateName,omitempty"`
	Labels          map[string]string `json:"labels"`
	ManagedFields   []any             `json:"managedFields,omitempty"`
	Namespace       string            `json:"namespace"`
	OwnerReferences []string          `json:"ownerReferences,omitempty"`
}

type OpenstackMachineResponse struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Metadata `json:"metadata"`
}

func (p *OpenstackMachine) Bytes() ([]byte, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (h *Helper) CreateOpenstackMachine(provider *OpenstackMachine) (*OpenstackMachineResponse, error) {
	u, err := url.Parse(h.Options.Url)
	if err != nil {
		return nil, err
	}

	u.Path = "/v1/rke-machine-config.cattle.io.openstackconfigs/fleet-default"
	b, err := provider.Bytes()
	if err != nil {
		return nil, err
	}

	opsMachineResp := &OpenstackMachineResponse{}
	resp, err := h.Http.R().
		SetResult(opsMachineResp).
		SetHeaders(GenAuthHeaders(h.Options.Token)).
		SetBody(string(b)).
		Post(u.String())
	if err != nil {
		return nil, err
	}

	if !resp.IsError() {
		return opsMachineResp, nil
	}

	return nil, fmt.Errorf(
		"failed to create openstack machine (%d %s)",
		resp.StatusCode(),
		resp.String(),
	)
}
