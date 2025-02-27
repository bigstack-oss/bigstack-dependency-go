package openstack

import (
	"os"

	"github.com/gophercloud/gophercloud/v2"
)

var (
	DefaultEndpointOpts = gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	}
)

type Options struct {
	ConfFile         string `json:"confFile" yaml:"confFile"`
	IdentityEndpoint string `json:"identityEndpoint" yaml:"identityEndpoint"`
	Auth             `json:"auth" yaml:"auth"`

	Domain  `json:"domain" yaml:"domain"`
	Tenant  `json:"tenant" yaml:"tenant"`
	Project `json:"project" yaml:"project"`
	User    `json:"user" yaml:"user"`

	Password string `json:"password" yaml:"password"`
	Passcode string `json:"passcode" yaml:"passcode"`

	IdentityAPIVersion string `json:"identityAPIVersion" yaml:"identityAPIVersion"`
	ImageAPIVersion    string `json:"imageAPIVersion" yaml:"imageAPIVersion"`

	Scope *gophercloud.AuthScope `json:"scope" yaml:"scope"`
}

type Auth struct {
	Source          string `json:"source" yaml:"source"`
	Type            string `json:"type" yaml:"type"`
	Url             string `json:"url" yaml:"url"`
	Username        string `json:"username" yaml:"username"`
	Password        string `json:"password" yaml:"password"`
	Token           string `json:"token" yaml:"token"`
	Project         `json:"project" yaml:"project"`
	EnableAutoRenew bool   `json:"enableAutoRenew" yaml:"enableAutoRenew"`
	File            string `json:"file" yaml:"file"`
}

type Tenant struct {
	ID     string `json:"id" yaml:"id"`
	Name   string `json:"name" yaml:"name"`
	Domain `json:"domain" yaml:"domain"`
}

type Project struct {
	ID     string `json:"id" yaml:"id"`
	Name   string `json:"name" yaml:"name"`
	Domain `json:"domain" yaml:"domain"`
}

type User struct {
	ID     string `json:"id" yaml:"id"`
	Name   string `json:"name" yaml:"name"`
	Domain `json:"domain" yaml:"domain"`
}

type Domain struct {
	ID   string `json:"id" yaml:"id"`
	Name string `json:"name" yaml:"name"`
}

func ConfFile(confFile string) Option {
	return func(o *Options) {
		o.ConfFile = confFile
	}
}

func AuthType(authType string) Option {
	return func(o *Options) {
		o.Auth.Type = authType
	}
}

func AuthUrl(AuthUrl string) Option {
	return func(o *Options) {
		o.Auth.Url = AuthUrl
	}
}

func UserID(userID string) Option {
	return func(o *Options) {
		o.User.ID = userID
	}
}

func Username(username string) Option {
	return func(o *Options) {
		o.User.Name = username
	}
}

func Password(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}

func Passcode(passcode string) Option {
	return func(o *Options) {
		o.Passcode = passcode
	}
}

func EnableAutoRenew(enableAutoRenew bool) Option {
	return func(o *Options) {
		o.Auth.EnableAutoRenew = enableAutoRenew
	}
}

func TenantID(tenantID string) Option {
	return func(o *Options) {
		o.Tenant.ID = tenantID
	}
}

func TenantName(tenantName string) Option {
	return func(o *Options) {
		o.Tenant.Name = tenantName
	}
}

func ProjectName(projectName string) Option {
	return func(o *Options) {
		o.Project.Name = projectName
	}
}

func DomainID(domainID string) Option {
	return func(o *Options) {
		o.Domain.ID = domainID
	}
}

func DomainName(domainName string) Option {
	return func(o *Options) {
		o.Domain.Name = domainName
	}
}

func ProjectDomainName(projectDomainName string) Option {
	return func(o *Options) {
		o.Project.Domain.Name = projectDomainName
	}
}

func UserDomainName(userDomainName string) Option {
	return func(o *Options) {
		o.User.Domain.Name = userDomainName
	}
}

func IdentityAPIVersion(identityAPIVersion string) Option {
	return func(o *Options) {
		o.IdentityAPIVersion = identityAPIVersion
	}
}

func ImageAPIVersion(imageAPIVersion string) Option {
	return func(o *Options) {
		o.ImageAPIVersion = imageAPIVersion
	}
}

func Scope(scope *gophercloud.AuthScope) Option {
	return func(o *Options) {
		o.Scope = scope
	}
}
