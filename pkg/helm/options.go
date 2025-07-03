package helm

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/kubernetes"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
)

var (
	Opts *Options
)

type Option func(*Options)

type Options struct {
	Release   string
	ChartPath string

	Chart           *chart.Chart
	Values          map[string]any
	CreateNamespace bool

	EnvConfig    *cli.EnvSettings
	ActionConfig *action.Configuration
	KubeConfig   kubernetes.Options
}

func Release(name string) Option {
	return func(o *Options) {
		o.Release = name
	}
}

func ChartPath(path string) Option {
	return func(o *Options) {
		o.ChartPath = path
	}
}

func CreateNamespace(enabled bool) Option {
	return func(o *Options) {
		o.CreateNamespace = enabled
	}
}

func EnvConfig(env *cli.EnvSettings) Option {
	return func(o *Options) {
		o.EnvConfig = env
	}
}

func ActionConfig(action *action.Configuration) Option {
	return func(o *Options) {
		o.ActionConfig = action
	}
}

func KubeConfig(kube kubernetes.Options) Option {
	return func(o *Options) {
		o.KubeConfig = kube
	}
}

func AuthType(authType string) Option {
	return func(o *Options) {
		o.KubeConfig.Auth.Type = authType
	}
}

func AuthFile(file string) Option {
	return func(o *Options) {
		o.KubeConfig.FilePath = file
	}
}
