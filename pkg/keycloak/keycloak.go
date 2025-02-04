package keycloak

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"

	"github.com/Nerzal/gocloak/v13"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/go-resty/resty/v2"
)

var (
	helper *Helper
	once   sync.Once
)

type Client interface {
	RestyClient() *resty.Client
	Login(context.Context, string, string, string, string, string) (*gocloak.JWT, error)
	LoginAdmin(context.Context, string, string, string) (*gocloak.JWT, error)
	GetUsers(context.Context, string, string, gocloak.GetUsersParams) ([]*gocloak.User, error)
	CreateClient(context.Context, string, string, gocloak.Client) (string, error)
	LogoutUserSession(context.Context, string, string, string) error
}

type Helper struct {
	Client
	Token string

	Options
}

func initOptions(opts []Option) *Options {
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	return options
}

func NewHelper(opts ...Option) (*Helper, error) {
	initedOpts := initOptions(opts)
	h := &Helper{Options: *initedOpts}

	err := h.SetKeycloakClient()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func NewGlobalHelper(opts ...Option) error {
	var err error
	once.Do(func() {
		helper, err = NewHelper(opts...)
		if err != nil {
			return
		}
	})
	if err != nil {
		return err
	}

	return nil
}

func GetGlobalHelper() *Helper {
	return helper
}

func (h *Helper) SetKeycloakClient() error {
	if h.Options.Host == "" {
		return fmt.Errorf("keycloak host is empty")
	}

	if h.Options.Username == "" {
		return fmt.Errorf("keycloak username is empty")
	}

	if h.Options.Password == "" {
		return fmt.Errorf("keycloak password is empty")
	}

	if h.Options.Realm == "" {
		return fmt.Errorf("keycloak realm is empty")
	}

	h.Client = gocloak.NewClient(h.Options.Host)
	return nil
}

func (h *Helper) LoginAdmin() error {
	if h.Options.TlsInsecureSkipVerify {
		h.Client.RestyClient().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	ctx, cancel := context.WithTimeout(wait.CtxMinutes(2))
	defer cancel()
	token, err := h.Client.LoginAdmin(
		ctx,
		h.Options.Username,
		h.Options.Password,
		h.Options.Realm,
	)
	if err == nil {
		h.Token = token.AccessToken
		return nil
	}

	return fmt.Errorf(
		"keycloak login failed: %s",
		err.Error(),
	)
}

func (h *Helper) LogoutUserSession(realm, sessionID string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.LogoutUserSession(ctx, h.Token, realm, sessionID)
}

func (h *Helper) CreateClient(realm string, opts gocloak.Client) (string, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.CreateClient(ctx, h.Token, realm, opts)
}
