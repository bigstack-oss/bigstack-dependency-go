package rancher

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type Repo struct {
	Type     string `json:"type"`
	Metadata `json:"metadata"`
	Spec     RepoSpec `json:"spec"`
}

type RepoSpec struct {
	Url                   string    `json:"url"`
	ClientSecret          SecretRef `json:"clientSecret"`
	InsecurePlainHttp     bool      `json:"insecurePlainHttp"`
	InsecureSkipTlsVerify bool      `json:"insecureSkipTLSVerify"`
}

type SecretRef struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type RepoResponse struct {
	Metadata `json:"metadata"`
}

func (r *Repo) Bytes() ([]byte, error) {
	return json.Marshal(r)
}

func (h *Helper) CreateClusterRepo(clusterId string, repo *Repo) (*RepoResponse, error) {
	u, err := url.Parse(h.Options.Url)
	if err != nil {
		return nil, err
	}

	u.Path = fmt.Sprintf("/k8s/clusters/%s/v1/catalog.cattle.io.clusterrepos", clusterId)
	b, err := repo.Bytes()
	if err != nil {
		return nil, err
	}

	repoResp := &RepoResponse{}
	resp, err := h.Http.R().
		SetResult(repoResp).
		SetHeaders(GenAuthHeaders(h.Options.Token)).
		SetBody(string(b)).
		Post(u.String())
	if err != nil {
		return nil, err
	}

	if !resp.IsError() {
		return repoResp, nil
	}

	return nil, fmt.Errorf(
		"failed to create cluster repo (%d %s)",
		resp.StatusCode(),
		resp.String(),
	)
}
