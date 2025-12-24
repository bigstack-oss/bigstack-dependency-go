package openstack

import (
	"bufio"
	"context"
	"os"
	"strings"
	"sync"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	log "go-micro.dev/v5/logger"
)

var (
	Opts   *Options
	helper *Helper

	once sync.Once
)

type Helper struct {
	Provider *gophercloud.ProviderClient

	Identity     *gophercloud.ServiceClient
	Compute      *gophercloud.ServiceClient
	Image        *gophercloud.ServiceClient
	Network      *gophercloud.ServiceClient
	Dns          *gophercloud.ServiceClient
	Loadbalancer *gophercloud.ServiceClient
	Storage      *gophercloud.ServiceClient
	Share        *gophercloud.ServiceClient
	ObjectStore  *gophercloud.ServiceClient

	*Options
}

type Option func(*Options)

func GetGlobalHelper() *Helper {
	return helper
}

func NewGlobalHelper(opts ...Option) error {
	var err error
	once.Do(func() {
		helper, err = NewHelper(opts...)
	})
	if err != nil {
		return err
	}

	return nil
}

func NewHelper(opts ...Option) (*Helper, error) {
	provider, err := newProvider(opts...)
	if err != nil {
		log.Errorf("failed to create provider: %s", err.Error())
		return nil, err
	}

	identityCli, err := newIdentityCli(provider)
	if err != nil {
		log.Errorf("failed to create identity client: %s", err.Error())
		return nil, err
	}

	computeCli, err := newComputeCli(provider)
	if err != nil {
		log.Errorf("failed to create compute client: %s", err.Error())
		return nil, err
	}

	imageCli, err := newImageCli(provider)
	if err != nil {
		log.Errorf("failed to create image client: %s", err.Error())
		return nil, err
	}

	networkCli, err := newNetworkCli(provider)
	if err != nil {
		log.Errorf("failed to create network client: %s", err.Error())
		return nil, err
	}

	dnsCli, err := newDnsCli(provider)
	if err != nil {
		log.Errorf("failed to create dns client: %s", err.Error())
		return nil, err
	}

	loadBalancerCli, err := newLoadBalancerCli(provider)
	if err != nil {
		log.Errorf("failed to create loadbalancer client: %s", err.Error())
		return nil, err
	}

	storageCli, err := newStorageCli(provider)
	if err != nil {
		log.Errorf("failed to create storage client: %s", err.Error())
		return nil, err
	}

	shareCli, err := newShareCli(provider)
	if err != nil {
		log.Errorf("failed to create share client: %s", err.Error())
		return nil, err
	}

	objectStoreCli, err := newObjectStoreCli(provider)
	if err != nil {
		log.Errorf("failed to create object store client: %s", err.Error())
		return nil, err
	}

	return &Helper{
		Provider:     provider,
		Identity:     identityCli,
		Compute:      computeCli,
		Image:        imageCli,
		Network:      networkCli,
		Dns:          dnsCli,
		Loadbalancer: loadBalancerCli,
		Storage:      storageCli,
		Share:        shareCli,
		ObjectStore:  objectStoreCli,
	}, nil
}

func newProvider(opts ...Option) (*gophercloud.ProviderClient, error) {
	syncedOpts, err := syncOptions(opts)
	if err != nil {
		return nil, err
	}

	finalOpts, err := GenAuthOpts(syncedOpts)
	if err != nil {
		return nil, err
	}

	return openstack.AuthenticatedClient(
		context.Background(),
		finalOpts,
	)
}

func syncOptions(opts []Option) (*Options, error) {
	options, err := NewConf()
	if err != nil {
		return nil, err
	}

	for _, o := range opts {
		o(options)
	}

	return options, nil
}

func GenAuthOpts(opts *Options) (gophercloud.AuthOptions, error) {
	if opts.Auth.Type == "env" {
		return ParseAuthEnv(opts)
	}

	if opts.Auth.Source == "file" {
		ParseAuthFile(opts)
	}

	return gophercloud.AuthOptions{
		IdentityEndpoint: opts.Auth.Url,
		Username:         opts.User.Name,
		Password:         opts.Password,
		TenantName:       opts.Project.Name,
		DomainName:       opts.Domain.Name,
		AllowReauth:      opts.EnableAutoRenew,
	}, nil
}

func ParseAuthEnv(opts *Options) (gophercloud.AuthOptions, error) {
	env, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return gophercloud.AuthOptions{}, err
	}

	env.AllowReauth = opts.EnableAutoRenew
	return env, nil
}

func ParseAuthFile(opts *Options) {
	openedFile, err := os.Open(opts.Auth.File)
	if err != nil {
		log.Errorf("failed to load ops conf: %s (%s)", opts.Auth.File, err.Error())
		return
	}
	defer openedFile.Close()
	s := bufio.NewScanner(openedFile)
	s.Split(bufio.ScanLines)

	for s.Scan() {
		switch {
		case strings.Contains(s.Text(), "OS_AUTH_URL"):
			words := strings.Split(s.Text(), "=")
			opts.Auth.Url = words[1]
		case strings.Contains(s.Text(), "OS_AUTH_TYPE"):
			words := strings.Split(s.Text(), "=")
			opts.Auth.Type = words[1]
		case strings.Contains(s.Text(), "OS_USERNAME"):
			words := strings.Split(s.Text(), "=")
			opts.User.Name = words[1]
		case strings.Contains(s.Text(), "OS_USER_DOMAIN_NAME"):
			words := strings.Split(s.Text(), "=")
			opts.User.Domain.Name = words[1]
		case strings.Contains(s.Text(), "OS_PASSWORD"):
			words := strings.Split(s.Text(), "=")
			opts.Password = words[1]
		case strings.Contains(s.Text(), "OS_PROJECT_NAME"):
			words := strings.Split(s.Text(), "=")
			opts.Project.Name = words[1]
		case strings.Contains(s.Text(), "OS_PROJECT_DOMAIN_NAME"):
			words := strings.Split(s.Text(), "=")
			opts.Project.Domain.Name = words[1]
		}
	}
}

func NewConf() (*Options, error) {
	opts := &Options{
		Domain: Domain{
			Name: "default",
		},
		Auth: Auth{
			Type: os.Getenv("OS_AUTH_TYPE"),
			Url:  os.Getenv("OS_AUTH_URL"),
		},
		User: User{
			Name: os.Getenv("OS_USERNAME"),
			Domain: Domain{
				Name: os.Getenv("OS_USER_DOMAIN_NAME"),
			},
		},
		Password: os.Getenv("OS_PASSWORD"),
		Tenant: Tenant{
			Name: os.Getenv("OS_PROJECT_NAME"),
			Domain: Domain{
				Name: os.Getenv("OS_PROJECT_DOMAIN_NAME"),
			},
		},
	}

	systemScope := os.Getenv("OS_SYSTEM_SCOPE")
	if systemScope == "all" {
		opts.Scope = &gophercloud.AuthScope{
			System: true,
		}
	}

	return opts, nil
}

func newIdentityCli(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	return openstack.NewIdentityV3(
		provider,
		gophercloud.EndpointOpts{
			Region: os.Getenv("OS_REGION_NAME"),
		},
	)
}

func newComputeCli(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	return openstack.NewComputeV2(
		provider,
		gophercloud.EndpointOpts{
			Region: os.Getenv("OS_REGION_NAME"),
		},
	)
}

func newImageCli(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	return openstack.NewImageV2(
		provider,
		gophercloud.EndpointOpts{
			Region: os.Getenv("OS_REGION_NAME"),
		},
	)
}

func newNetworkCli(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	return openstack.NewNetworkV2(
		provider,
		gophercloud.EndpointOpts{
			Region: os.Getenv("OS_REGION_NAME"),
		},
	)
}

func newDnsCli(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	return openstack.NewDNSV2(
		provider,
		gophercloud.EndpointOpts{
			Region: os.Getenv("OS_REGION_NAME"),
		},
	)
}

func newLoadBalancerCli(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	return openstack.NewLoadBalancerV2(
		provider,
		gophercloud.EndpointOpts{
			Region: os.Getenv("OS_REGION_NAME"),
		},
	)
}

func newStorageCli(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	return openstack.NewBlockStorageV3(
		provider,
		gophercloud.EndpointOpts{
			Region:  os.Getenv("OS_REGION_NAME"),
			Version: 3,
		},
	)
}

func newShareCli(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	return openstack.NewSharedFileSystemV2(
		provider,
		gophercloud.EndpointOpts{
			Region: os.Getenv("OS_REGION_NAME"),
		},
	)
}

func newObjectStoreCli(provider *gophercloud.ProviderClient) (*gophercloud.ServiceClient, error) {
	return openstack.NewObjectStorageV1(
		provider,
		gophercloud.EndpointOpts{
			Region: os.Getenv("OS_REGION_NAME"),
		},
	)
}
