package helm

import "helm.sh/helm/v3/pkg/cli/values"

type Chart struct {
	Release             string `yaml:"release"`
	Version             string `yaml:"version"`
	Namespace           string `yaml:"namespace"`
	Tgz                 `yaml:"tgz"`
	CustomizedValues    *values.Options `yaml:"customizedValues"`
	ClusterRolesToPatch []string        `yaml:"clusterRolesToPatch"`
}

type Tgz struct {
	Remote string `yaml:"remote"`
	Local  string `yaml:"local"`
}
