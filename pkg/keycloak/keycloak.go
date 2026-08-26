package keycloak

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"
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
	LoginClient(ctx context.Context, clientID, clientSecret, realm string, scopes ...string) (*gocloak.JWT, error)
	GetUsers(context.Context, string, string, gocloak.GetUsersParams) ([]*gocloak.User, error)
	GetClients(context.Context, string, string, gocloak.GetClientsParams) ([]*gocloak.Client, error)
	CreateClient(context.Context, string, string, gocloak.Client) (string, error)
	CreateClientProtocolMapper(context.Context, string, string, string, gocloak.ProtocolMapperRepresentation) (string, error)
	GetClientSecret(ctx context.Context, token, realm, idOfClient string) (*gocloak.CredentialRepresentation, error)
	CreateUser(context.Context, string, string, gocloak.User) (string, error)
	SetPassword(context.Context, string, string, string, string, bool) error
	UpdateUser(context.Context, string, string, gocloak.User) error
	DeleteUser(context.Context, string, string, string) error
	GetClientRole(ctx context.Context, token string, realm string, idOfClient string, roleName string) (*gocloak.Role, error)
	AddClientRolesToUser(ctx context.Context, token string, realm string, idOfClient string, userID string, roles []gocloak.Role) error
	DeleteClientRolesFromUser(ctx context.Context, token string, realm string, idOfClient string, userID string, roles []gocloak.Role) error

	/* Client mirrors methods gocloak.GoCloak already implements -- there's no
	 * implementation here because gocloak's real client satisfies this interface
	 * directly (assigned in SetKeycloakClient). This exists purely so mockery can
	 * generate a mock for tests.
	 */
	GetGroups(ctx context.Context, token string, realm string, params gocloak.GetGroupsParams) ([]*gocloak.Group, error)
	GetGroup(ctx context.Context, token string, realm string, groupID string) (*gocloak.Group, error)
	CreateGroup(ctx context.Context, token string, realm string, group gocloak.Group) (string, error)
	CreateChildGroup(ctx context.Context, token string, realm string, groupID string, group gocloak.Group) (string, error)
	GetUserGroups(ctx context.Context, token string, realm string, userID string, params gocloak.GetGroupsParams) ([]*gocloak.Group, error)
	AddUserToGroup(ctx context.Context, token string, realm string, userID string, groupID string) error
	DeleteUserFromGroup(ctx context.Context, token string, realm string, userID string, groupID string) error
	GetClientRolesByUserID(ctx context.Context, token string, realm string, idOfClient string, userID string) ([]*gocloak.Role, error)
	GetRealmRole(ctx context.Context, token string, realm string, roleName string) (*gocloak.Role, error)
	AddRealmRoleToUser(ctx context.Context, token string, realm string, userID string, roles []gocloak.Role) error
	LogoutUserSession(context.Context, string, string, string) error
	GetComponent(ctx context.Context, token string, realm string, componentID string) (*gocloak.Component, error)
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

	hasUserCreds := h.Options.Username != "" && h.Options.Password != ""
	hasServiceAccountCreds := h.Options.ClientID != "" && h.Options.ClientSecret != ""
	if !hasUserCreds && !hasServiceAccountCreds {
		return fmt.Errorf("keycloak credentials are empty: need either username/password or clientId/clientSecret")
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

/*
 * LoginServiceAccount authenticates as a confidential client's service account
 * (client_credentials grant) instead of a human admin's username/password.
 * Unlike LoginAdmin, this has no separate expiry to manage beyond the resulting
 * access token itself: the client secret backing it doesn't expire on its own,
 * so a fresh token is always a re-login away.
 */
func (h *Helper) LoginServiceAccount() error {
	if h.Options.TlsInsecureSkipVerify {
		h.Client.RestyClient().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	ctx, cancel := context.WithTimeout(wait.CtxMinutes(2))
	defer cancel()
	token, err := h.Client.LoginClient(
		ctx,
		h.Options.ClientID,
		h.Options.ClientSecret,
		h.Options.Realm,
	)
	if err == nil {
		h.Token = token.AccessToken
		return nil
	}

	return fmt.Errorf(
		"keycloak service account login failed: %s",
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

func (h *Helper) DeleteUser(realm, userID string) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.DeleteUser(ctx, h.Token, realm, userID)
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

func (h *Helper) DeleteClientRolesFromUser(realm, clientId, userID string, roles []gocloak.Role) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.DeleteClientRolesFromUser(ctx, h.Token, realm, clientId, userID, roles)
}

func (h *Helper) GetRealmRole(realm, roleName string) (*gocloak.Role, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.GetRealmRole(ctx, h.Token, realm, roleName)
}

func (h *Helper) AddRealmRoleToUser(realm, userID string, roles []gocloak.Role) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.AddRealmRoleToUser(ctx, h.Token, realm, userID, roles)
}

func (h *Helper) GetComponent(realm, componentID string) (*gocloak.Component, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.GetComponent(ctx, h.Token, realm, componentID)
}

// Without this, an unset Max silently inherits Keycloak's server-side default
// page size, truncating a realm/user with more groups than that.
const groupsPageSize = 100

func paginateGroups(fetch func(gocloak.GetGroupsParams) ([]*gocloak.Group, error), params gocloak.GetGroupsParams) ([]*gocloak.Group, error) {
	if params.Max != nil {
		return fetch(params)
	}

	var all []*gocloak.Group
	first := 0
	for {
		page := params
		page.First = gocloak.IntP(first)
		page.Max = gocloak.IntP(groupsPageSize)

		got, err := fetch(page)
		if err != nil {
			return nil, err
		}

		all = append(all, got...)
		if len(got) < groupsPageSize {
			return all, nil
		}
		first += groupsPageSize
	}
}

func (h *Helper) GetGroups(realm string, params gocloak.GetGroupsParams) ([]*gocloak.Group, error) {
	return paginateGroups(func(p gocloak.GetGroupsParams) ([]*gocloak.Group, error) {
		ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
		defer cancel()
		return h.Client.GetGroups(ctx, h.Token, realm, p)
	}, params)
}

func (h *Helper) GetUserGroups(realm, userID string) ([]*gocloak.Group, error) {
	return paginateGroups(func(p gocloak.GetGroupsParams) ([]*gocloak.Group, error) {
		ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
		defer cancel()
		return h.Client.GetUserGroups(ctx, h.Token, realm, userID, p)
	}, gocloak.GetGroupsParams{})
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

func (h *Helper) GetOrCreateGroupPath(realm, path string) (*gocloak.Group, error) {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	for _, s := range segments {
		if s == "" {
			return nil, fmt.Errorf("invalid group path %q", path)
		}
	}

	var current *gocloak.Group
	for _, name := range segments {
		child, err := h.findChildGroup(realm, current, name)
		if err != nil {
			return nil, err
		}
		if child == nil {
			child, err = h.createChildGroup(realm, current, name)
			if err != nil {
				return nil, err
			}
		}
		current = child
	}

	return current, nil
}

func (h *Helper) findChildGroup(realm string, parent *gocloak.Group, name string) (*gocloak.Group, error) {
	if parent == nil {
		groups, err := h.GetGroups(realm, gocloak.GetGroupsParams{})
		if err != nil {
			return nil, err
		}
		for _, group := range groups {
			if group.Name != nil && *group.Name == name {
				return group, nil
			}
		}
		return nil, nil
	}

	// GET /groups/{id}/children 405s on 17.0.1-legacy -- children only show up nested under
	// the parent's own GET response. This pins the whole helper to that legacy Group
	// representation: Keycloak 23+ stops populating SubGroups here (subGroupCount + a working
	// /children endpoint instead), which this gocloak version has no model for, so on such a
	// server this reads as "no children" and GetOrCreateGroupPath will try to recreate an existing
	// group (409) instead of finding it.
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	fresh, err := h.Client.GetGroup(ctx, h.Token, realm, *parent.ID)
	if err != nil {
		return nil, err
	}
	if fresh.SubGroups == nil {
		return nil, nil
	}
	for _, subGroup := range *fresh.SubGroups {
		if subGroup.Name != nil && *subGroup.Name == name {
			return &subGroup, nil
		}
	}
	return nil, nil
}

func (h *Helper) createChildGroup(realm string, parent *gocloak.Group, name string) (*gocloak.Group, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()

	group := gocloak.Group{Name: &name}
	var id string
	var err error
	var path string
	if parent == nil {
		id, err = h.Client.CreateGroup(ctx, h.Token, realm, group)
		path = "/" + name
	} else {
		id, err = h.Client.CreateChildGroup(ctx, h.Token, realm, *parent.ID, group)
		path = *parent.Path + "/" + name
	}
	if err != nil {
		return nil, err
	}

	return &gocloak.Group{ID: &id, Name: &name, Path: &path}, nil
}

func (h *Helper) HasClientRole(realm, clientID, userID, roleName string) (bool, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	roles, err := h.Client.GetClientRolesByUserID(ctx, h.Token, realm, clientID, userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.Name != nil && *role.Name == roleName {
			return true, nil
		}
	}

	return false, nil
}
