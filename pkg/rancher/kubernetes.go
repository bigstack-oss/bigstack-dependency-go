package rancher

import (
	"encoding/json"
	"fmt"
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
	Configs map[string]string   `json:"configs"`
	Mirrors map[string]MirrorTo `json:"mirrors"`
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

func (h *Helper) WaitKubernetesActive(name string) (*StatusResponse, error) {
	u, err := url.Parse(h.Options.Url)
	if err != nil {
		log.Errorf("rancher: failed to parse url(%v)", err)
		return nil, err
	}

	u.Path = fmt.Sprintf("/apis/provisioning.cattle.io/v1/namespaces/fleet-default/clusters/%s", name)
	attemptsMax := 240

	for {
		if attemptsMax <= 0 {
			break
		}

		wait.Seconds(10)
		statusResp := &StatusResponse{}
		resp, err := h.Http.R().
			SetResult(statusResp).
			SetHeaders(GenAuthHeaders(h.Options.Token)).
			Get(u.String())
		if err != nil {
			attemptsMax--
			log.Errorf("rancher: failed to request GET kubernetes status(%v)", err)
			continue
		}

		if !resp.IsError() {
			attemptsMax--
			log.Infof("rancher: has error response when requesting kubernetes status(%d %s)", resp.StatusCode(), resp.String())
			continue
		}

		if !http.Is2XXCode[resp.StatusCode()] {
			attemptsMax--
			log.Infof("rancher: kubernetes status response is not 2xx code(%d %s)", resp.StatusCode(), resp.String())
			continue
		}

		if statusResp.Status.Ready {
			return statusResp, nil
		}

		attemptsMax--
	}

	return nil, fmt.Errorf(
		"kubernetes cluster is not ready until %d seconds",
		10*240,
	)
}

func (h *Helper) GetKubernetesConfig(name string) ([]byte, error) {
	u, err := url.Parse(h.Options.Url)
	if err != nil {
		return nil, err
	}

	u.Path = fmt.Sprintf("/v3/clusters/%s", name)
	u.RawQuery = url.Values{"action": []string{"generateKubeconfig"}}.Encode()
	resp, err := h.Http.R().
		SetHeaders(GenAuthHeaders(h.Options.Token)).
		Post(u.String())
	if err != nil {
		return nil, err
	}

	if !resp.IsError() {
		return resp.Body(), nil
	}

	rawConf := map[string]any{}
	err = yaml.Unmarshal(resp.Body(), &rawConf)
	if err != nil {
		return nil, err
	}

	conf, found := rawConf["config"]
	if !found {
		return nil, fmt.Errorf("failed to find rke cluster config(%s)", name)
	}

	bytes, err := yaml.Marshal(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal rke cluster conf(%v)", err)
	}

	return bytes, nil
}
