package rancher

import (
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"net/url"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	log "go-micro.dev/v5/logger"
	"gopkg.in/yaml.v2"
)

type Cluster struct {
	Type     string `json:"type"`
	Metadata `json:"metadata"`
	Spec     `json:"spec"`
}

type ListClusterResponse struct {
	Data ListClusterData `json:"data"`
}

type ListClusterData struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	NodeCount int    `json:"nodeCount"`
	State     string `json:"state"`
	Version   `json:"version"`
	Labels    map[string]string `json:"labels"`
	Created   string            `json:"created"`
}

type Version struct {
	GitVersion string `json:"gitVersion"`
}

type Spec struct {
	RkeConfig                                            `json:"rkeConfig"`
	MachineSelectorConfig                                []MachineSelectorConfig `json:"machineSelectorConfig"`
	KubernetesVersion                                    string                  `json:"kubernetesVersion"`
	DefaultPodSecurityPolicyTemplateName                 string                  `json:"defaultPodSecurityPolicyTemplateName"`
	DefaultPodSecurityAdmissionConfigurationTemplateName string                  `json:"defaultPodSecurityAdmissionConfigurationTemplateName"`
	CloudCredentialSecretName                            string                  `json:"cloudCredentialSecretName"`
	LocalClusterAuthEndpoint                             `json:"localClusterAuthEndpoint"`
}

type RkeConfig struct {
	ChartValues           `json:"chartValues"`
	UpgradeStrategy       `json:"upgradeStrategy"`
	DataDirectories       `json:"dataDirectories"`
	MachineGlobalConfig   `json:"machineGlobalConfig"`
	MachineSelectorConfig []MachineSelectorConfig `json:"machineSelectorConfig"`
	Etcd                  `json:"etcd"`
	Registries            `json:"registries"`
	MachinePools          []MachinePool `json:"machinePools"`
}

type DataDirectories struct {
	SystemAgent  string `json:"systemAgent"`
	Provisioning string `json:"provisioning"`
	K8sDistro    string `json:"k8sDistro"`
}

type ChartValues struct {
	Rke2Cilium `json:"rke2-cilium"`
}

type Rke2Cilium struct {
	Cilium `json:"cilium,omitempty"`
}

type Cilium struct {
	Ipv6 `json:"ipv6"`
}

type Ipv6 struct {
	Enabled bool `json:"enabled"`
}

type UpgradeStrategy struct {
	ControlPlaneConcurrency  string `json:"controlPlaneConcurrency"`
	ControlPlaneDrainOptions `json:"controlPlaneDrainOptions"`
	WorkerConcurrency        string `json:"workerConcurrency"`
	WorkerDrainOptions       `json:"workerDrainOptions"`
}

type ControlPlaneDrainOptions struct {
	DeleteEmptyDirData           bool `json:"deleteEmptyDirData"`
	DisableEviction              bool `json:"disableEviction"`
	Enabled                      bool `json:"enabled"`
	Force                        bool `json:"force"`
	GracePeriod                  int  `json:"gracePeriod"`
	IgnoreDaemonSets             bool `json:"ignoreDaemonSets"`
	SkipWaitForDeleteTimeoutSecs int  `json:"skipWaitForDeleteTimeoutSeconds"`
	Timeout                      int  `json:"timeout"`
}

type WorkerDrainOptions struct {
	DeleteEmptyDirData           bool `json:"deleteEmptyDirData"`
	DisableEviction              bool `json:"disableEviction"`
	Enabled                      bool `json:"enabled"`
	Force                        bool `json:"force"`
	GracePeriod                  int  `json:"gracePeriod"`
	IgnoreDaemonSets             bool `json:"ignoreDaemonSets"`
	SkipWaitForDeleteTimeoutSecs int  `json:"skipWaitForDeleteTimeoutSeconds"`
	Timeout                      int  `json:"timeout"`
}

type MachineGlobalConfig struct {
	Cni               string `json:"cni"`
	DisableKubeProxy  bool   `json:"disable-kube-proxy"`
	EtcdExposeMetrics bool   `json:"etcd-expose-metrics"`
}

type MachineSelectorConfig struct {
	Config `json:"config"`
}

type Config struct {
	ProtectKernelDefaults bool `json:"protect-kernel-defaults,omitempty"`
}

type Etcd struct {
	DisableSnapshots     bool `json:"disableSnapshots"`
	*S3                  `json:"s3"`
	SnapshotRetention    int    `json:"snapshotRetention"`
	SnapshotScheduleCron string `json:"snapshotScheduleCron"`
}

type S3 struct{}

type Registries struct {
	Configs map[string]Registry `json:"configs"`
	Mirrors map[string]MirrorTo `json:"mirrors"`
}

type Registry struct {
	AuthConfigSecretName string `json:"authConfigSecretName"`
	CaBundle             string `json:"caBundle"`
	InsecureSkipVerify   bool   `json:"insecureSkipVerify"`
}

type MirrorTo struct {
	Endpoint []string          `json:"endpoint"`
	Rewrite  map[string]string `json:"rewrite"`
}

type MachinePool struct {
	Name              string `json:"name"`
	EtcdRole          bool   `json:"etcdRole"`
	ControlPlaneRole  bool   `json:"controlPlaneRole"`
	WorkerRole        bool   `json:"workerRole"`
	HostnamePrefix    string `json:"hostnamePrefix"`
	Labels            `json:"labels"`
	Quantity          int    `json:"quantity"`
	UnhealthyNodeTime string `json:"unhealthyNodeTimeout"`
	MachineConfigRef  `json:"machineConfigRef"`
	DrainBeforeDelete bool `json:"drainBeforeDelete"`
}

type Labels struct{}

type MachineConfigRef struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type LocalClusterAuthEndpoint struct {
	Enabled bool   `json:"enabled"`
	CaCerts string `json:"caCerts"`
	Fqdn    string `json:"fqdn"`
}

type ClusterResponse struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Metadata `json:"metadata"`
}

type StatusResponse struct {
	Kind   string `json:"kind"`
	Status `json:"status"`
}

type Status struct {
	ClusterName   string `json:"clusterName"`
	AgentDeployed bool   `json:"agentDeployed"`
	Ready         bool   `json:"ready"`
}

func (c *Cluster) Bytes() ([]byte, error) {
	return json.Marshal(c)
}

func (h *Helper) CreateRancherSecret(secret *Secret) (*SecretResponse, error) {
	u, err := url.Parse(h.Options.Url)
	if err != nil {
		return nil, err
	}

	u.Path = "/v1/secrets/fleet-default"
	b, err := secret.Bytes()
	if err != nil {
		return nil, err
	}

	secretResp := &SecretResponse{}
	resp, err := h.Http.R().
		SetResult(secretResp).
		SetHeaders(GenAuthHeaders(h.Options.Token)).
		SetBody(string(b)).
		Post(u.String())
	if err != nil {
		return nil, err
	}

	if !resp.IsError() {
		return secretResp, nil
	}

	return nil, fmt.Errorf(
		"failed to create secret (%d %s)",
		resp.StatusCode(),
		resp.String(),
	)
}

func (h *Helper) CreateClusterSecret(clusterId string, secret *Secret) (*SecretResponse, error) {
	u, err := url.Parse(h.Options.Url)
	if err != nil {
		return nil, err
	}

	u.Path = fmt.Sprintf("/k8s/clusters/%s/v1/secrets/cattle-system", clusterId)
	b, err := secret.Bytes()
	if err != nil {
		return nil, err
	}

	secretResp := &SecretResponse{}
	resp, err := h.Http.R().
		SetResult(secretResp).
		SetHeaders(GenAuthHeaders(h.Options.Token)).
		SetBody(string(b)).
		Post(u.String())
	if err != nil {
		return nil, err
	}

	if !resp.IsError() {
		return secretResp, nil
	}

	return nil, fmt.Errorf(
		"failed to create secret (%d %s)",
		resp.StatusCode(),
		resp.String(),
	)
}

func (h *Helper) CreateKubernetes(cluster *Cluster) (*ClusterResponse, error) {
	u, err := url.Parse(h.Options.Url)
	if err != nil {
		return nil, err
	}

	u.Path = "/v1/provisioning.cattle.io.clusters"
	b, err := cluster.Bytes()
	if err != nil {
		return nil, err
	}

	clusterResp := &ClusterResponse{}
	resp, err := h.Http.R().
		SetResult(clusterResp).
		SetHeaders(GenAuthHeaders(h.Options.Token)).
		SetBody(string(b)).
		Post(u.String())
	if err != nil {
		return nil, err
	}

	if !resp.IsError() {
		return clusterResp, nil
	}

	return nil, fmt.Errorf(
		"failed to create create kubernetes (%d %s)",
		resp.StatusCode(),
		resp.String(),
	)
}

func (h *Helper) ListKubernetes() (*ListClusterResponse, error) {
	u, err := url.Parse(h.Options.Url)
	if err != nil {
		return nil, err
	}

	u.Path = "/v3/clusters"
	clusters := &ListClusterResponse{}
	resp, err := h.Http.R().
		SetResult(clusters).
		SetHeaders(GenAuthHeaders(h.Options.Token)).
		Get(u.String())
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, err
	}

	return clusters, nil
}

func (h *Helper) DeleteKubernetes(name string) error {
	u, err := url.Parse(h.Options.Url)
	if err != nil {
		return err
	}

	u.Path = fmt.Sprintf("/v1/provisioning.cattle.io.clusters/fleet-default/%s", name)
	resp, err := h.Http.R().
		SetHeaders(GenAuthHeaders(h.Options.Token)).
		Delete(u.String())
	if err != nil {
		return err
	}
	if !resp.IsError() {
		return nil
	}

	return fmt.Errorf(
		"failed to delete kubernetes %s (%d %s)",
		name,
		resp.StatusCode(),
		string(resp.Body()),
	)
}

func (h *Helper) WaitKubernetesActive(name string) (*StatusResponse, error) {
	u, err := url.Parse(h.Options.Url)
	if err != nil {
		log.Errorf("rancher: failed to parse url(%v)", err)
		return nil, err
	}

	u.Path = fmt.Sprintf("/apis/provisioning.cattle.io/v1/namespaces/fleet-default/clusters/%s", name)
	attemptsMax := 240
	for range attemptsMax {
		wait.Seconds(10)
		statusResp := &StatusResponse{}
		resp, err := h.Http.R().
			SetResult(statusResp).
			SetHeaders(GenAuthHeaders(h.Options.Token)).
			Get(u.String())
		if err != nil {
			log.Errorf("rancher: failed to request GET kubernetes status(%v)", err)
			continue
		}

		if !http.Is2XXCode[resp.StatusCode()] {
			log.Infof("rancher: kubernetes status response is not 2xx code(%d %s)", resp.StatusCode(), resp.String())
			continue
		}

		if statusResp.Status.Ready {
			return statusResp, nil
		}

		log.Infof("rancher: kubernetes cluster %s is not ready yet, waiting...", name)
	}

	return nil, fmt.Errorf(
		"kubernetes cluster is not ready until %d seconds",
		10*240,
	)
}

func (h *Helper) WaitKubernetesDeleted(name string) error {
	u, err := url.Parse(h.Options.Url)
	if err != nil {
		return err
	}

	u.Path = fmt.Sprintf("/apis/provisioning.cattle.io/v1/namespaces/fleet-default/clusters/%s", name)
	attemptsMax := 240
	for range attemptsMax {
		wait.Seconds(10)
		resp, err := h.Http.R().
			SetHeaders(GenAuthHeaders(h.Options.Token)).
			Get(u.String())
		if err != nil {
			continue
		}

		if resp.StatusCode() == nethttp.StatusNotFound {
			return nil
		}
	}

	return fmt.Errorf(
		"failed to delete kubernetes %s until %d seconds",
		name,
		10*240,
	)
}

func (h *Helper) GetKubernetesConfig(name string) ([]byte, error) {
	u, err := url.Parse(h.Options.Url)
	if err != nil {
		log.Errorf("rancher: failed to parse url(%v)", err)
		return nil, err
	}

	u.Path = fmt.Sprintf("/v3/clusters/%s", name)
	u.RawQuery = url.Values{"action": []string{"generateKubeconfig"}}.Encode()
	resp, err := h.Http.R().
		SetHeaders(GenAuthHeaders(h.Options.Token)).
		Post(u.String())
	if err != nil {
		log.Errorf("rancher: failed to request generate kubernetes config(%v)", err)
		return nil, err
	}

	if resp.IsError() {
		err := fmt.Errorf("failed to generate kubernetes config(%d %s)", resp.StatusCode(), resp.String())
		log.Errorf("rancher: %v", err)
		return nil, err
	}

	rawConf := map[string]any{}
	err = yaml.Unmarshal(resp.Body(), &rawConf)
	if err != nil {
		log.Errorf("rancher: failed to unmarshal kubernetes config(%v)", err)
		return nil, err
	}

	conf, found := rawConf["config"]
	if !found {
		err := fmt.Errorf("failed to find rke cluster config(%s)", name)
		log.Errorf("rancher: %v", err)
		return nil, err
	}

	bytes, err := yaml.Marshal(conf)
	if err != nil {
		log.Errorf("rancher: failed to marshal rke cluster conf(%v)", err)
		return nil, err
	}

	return bytes, nil
}
