package keycloak

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
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
	GetClients(context.Context, string, string, gocloak.GetClientsParams) ([]*gocloak.Client, error)
	CreateClient(context.Context, string, string, gocloak.Client) (string, error)
	CreateClientProtocolMapper(context.Context, string, string, string, gocloak.ProtocolMapperRepresentation) (string, error)
	GetClientSecret(ctx context.Context, token, realm, idOfClient string) (*gocloak.CredentialRepresentation, error)
	CreateUser(context.Context, string, string, gocloak.User) (string, error)
	SetPassword(context.Context, string, string, string, string, bool) error
	UpdateUser(context.Context, string, string, gocloak.User) error
	GetClientRole(ctx context.Context, token string, realm string, idOfClient string, roleName string) (*gocloak.Role, error)
	AddClientRolesToUser(ctx context.Context, token string, realm string, idOfClient string, userID string, roles []gocloak.Role) error
	GetGroups(ctx context.Context, token string, realm string, params gocloak.GetGroupsParams) ([]*gocloak.Group, error)
	AddUserToGroup(ctx context.Context, token string, realm string, userID string, groupID string) error
	DeleteUserFromGroup(ctx context.Context, token string, realm string, userID string, groupID string) error
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
	if h.Options.Scheme == "" {
		return fmt.Errorf("keycloak scheme is empty")
	}

	if h.Options.Ip == "" {
		return fmt.Errorf("keycloak ip is empty")
	}

	if h.Options.Port == 0 {
		return fmt.Errorf("keycloak port is empty")
	}

	if h.Options.Path == "" {
		return fmt.Errorf("keycloak path is empty")
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

	url := h.genKeycloakUrl()
	h.Client = gocloak.NewClient(url)
	return nil
}

func (h *Helper) genKeycloakUrl() string {
	u := url.URL{}
	u.Scheme = h.Options.Scheme
	u.Host = fmt.Sprintf("%s:%d", h.Options.Ip, h.Options.Port)
	u.Path = h.Options.Path
	return u.String()
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

func (h *Helper) GetClients(realm string, params gocloak.GetClientsParams) ([]*gocloak.Client, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.GetClients(ctx, h.Token, realm, params)
}

func (h *Helper) CreateClientProtocolMapper(realm, clientId string, opts gocloak.ProtocolMapperRepresentation) (string, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.CreateClientProtocolMapper(ctx, h.Token, realm, clientId, opts)
}

func (h *Helper) GetClientSecret(realm, clientId string) (*gocloak.CredentialRepresentation, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.GetClientSecret(ctx, h.Token, realm, clientId)
}

func (h *Helper) CreateUser(realm string, user gocloak.User) (string, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.CreateUser(ctx, h.Token, realm, user)
}

func (h *Helper) GetUser(realm, name string) (*gocloak.User, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	users, err := h.Client.GetUsers(ctx, h.Token, realm, gocloak.GetUsersParams{})
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Username != nil && *user.Username == name {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user %s not found", name)
}

func (h *Helper) SetPassword(realm, userID, password string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.SetPassword(ctx, h.Token, userID, realm, password, false)
}

func (h *Helper) UpdateUser(realm string, user gocloak.User) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.UpdateUser(ctx, h.Token, realm, user)
}

func (h *Helper) GetClientRole(realm, clientId, roleName string) (*gocloak.Role, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.GetClientRole(ctx, h.Token, realm, clientId, roleName)
}

func (h *Helper) AddClientRolesToUser(realm, clientId, userID string, roles []gocloak.Role) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.AddClientRolesToUser(ctx, h.Token, realm, clientId, userID, roles)
}

func (h *Helper) GetGroups(realm string, params gocloak.GetGroupsParams) ([]*gocloak.Group, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.GetGroups(ctx, h.Token, realm, params)
}

func (h *Helper) GetGroup(realm, name string) (*gocloak.Group, error) {
	groups, err := h.GetGroups(realm, gocloak.GetGroupsParams{})
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		if group.Name != nil && *group.Name == name {
			return group, nil
		}
	}

	return nil, fmt.Errorf("group %s not found", name)
}

func (h *Helper) AddUserToGroup(realm, userID, groupID string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.AddUserToGroup(ctx, h.Token, realm, userID, groupID)
}

func (h *Helper) DeleteUserFromGroup(realm, userID, groupID string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.DeleteUserFromGroup(ctx, h.Token, realm, userID, groupID)
}
