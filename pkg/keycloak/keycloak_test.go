package keycloak

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }

func groupNamed(id, name string) *gocloak.Group {
	return &gocloak.Group{ID: strPtr(id), Name: strPtr(name)}
}

func groupAt(id, name, path string) *gocloak.Group {
	return &gocloak.Group{ID: strPtr(id), Name: strPtr(name), Path: strPtr(path)}
}

func TestGetOrCreateGroupPath_InvalidPath(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "empty path", path: ""},
		{name: "empty interior segment", path: "cmp//admin"},
		{name: "multiple empty interior segments", path: "cmp///admin"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := &Helper{Client: NewMockClient(t)}
			_, err := h.GetOrCreateGroupPath("master", tc.path)
			require.EqualError(t, err, fmt.Sprintf("invalid group path %q", tc.path))
		})
	}
}

func TestGetOrCreateGroupPath(t *testing.T) {
	errBoom := errors.New("boom")

	tests := []struct {
		name          string
		path          string
		mockSetup     func(client *MockClient)
		expected      *gocloak.Group
		expectedError error
	}{
		{
			name: "Should return the existing group for a single-segment path found at the top level",
			path: "cmp",
			mockSetup: func(client *MockClient) {
				client.On("GetGroups", mock.Anything, mock.Anything, "master", mock.Anything).
					Return([]*gocloak.Group{groupAt("cmp-id", "cmp", "/cmp")}, nil)
			},
			expected: groupAt("cmp-id", "cmp", "/cmp"),
		},
		{
			name: "Should create a single-segment path not found at the top level, with Path populated",
			path: "cmp",
			mockSetup: func(client *MockClient) {
				client.On("GetGroups", mock.Anything, mock.Anything, "master", mock.Anything).
					Return([]*gocloak.Group{}, nil)
				client.On("CreateGroup", mock.Anything, mock.Anything, "master", mock.Anything).
					Return("cmp-id", nil)
			},
			expected: groupAt("cmp-id", "cmp", "/cmp"),
		},
		{
			name: "Should walk every segment of a multi-segment path that already exists",
			path: "cmp/PROJ001/admin",
			mockSetup: func(client *MockClient) {
				client.On("GetGroups", mock.Anything, mock.Anything, "master", mock.Anything).
					Return([]*gocloak.Group{groupAt("cmp-id", "cmp", "/cmp")}, nil)
				client.On("GetGroup", mock.Anything, mock.Anything, "master", "cmp-id").
					Return(&gocloak.Group{
						ID: strPtr("cmp-id"), Name: strPtr("cmp"), Path: strPtr("/cmp"),
						SubGroups: &[]gocloak.Group{*groupAt("proj-id", "PROJ001", "/cmp/PROJ001")},
					}, nil)
				client.On("GetGroup", mock.Anything, mock.Anything, "master", "proj-id").
					Return(&gocloak.Group{
						ID: strPtr("proj-id"), Name: strPtr("PROJ001"), Path: strPtr("/cmp/PROJ001"),
						SubGroups: &[]gocloak.Group{*groupAt("admin-id", "admin", "/cmp/PROJ001/admin")},
					}, nil)
			},
			expected: groupAt("admin-id", "admin", "/cmp/PROJ001/admin"),
		},
		{
			name: "Should create only the missing middle segment of a multi-segment path, with Path derived from its parent",
			path: "cmp/PROJ001/admin",
			mockSetup: func(client *MockClient) {
				client.On("GetGroups", mock.Anything, mock.Anything, "master", mock.Anything).
					Return([]*gocloak.Group{groupAt("cmp-id", "cmp", "/cmp")}, nil)
				client.On("GetGroup", mock.Anything, mock.Anything, "master", "cmp-id").
					Return(&gocloak.Group{ID: strPtr("cmp-id"), Name: strPtr("cmp"), Path: strPtr("/cmp"), SubGroups: &[]gocloak.Group{}}, nil)
				client.On("CreateChildGroup", mock.Anything, mock.Anything, "master", "cmp-id", mock.Anything).
					Return("proj-id", nil)
				client.On("GetGroup", mock.Anything, mock.Anything, "master", "proj-id").
					Return(&gocloak.Group{
						ID: strPtr("proj-id"), Name: strPtr("PROJ001"), Path: strPtr("/cmp/PROJ001"),
						SubGroups: &[]gocloak.Group{*groupAt("admin-id", "admin", "/cmp/PROJ001/admin")},
					}, nil)
			},
			expected: groupAt("admin-id", "admin", "/cmp/PROJ001/admin"),
		},
		{
			name: "Should stop and propagate an error finding the top-level group",
			path: "cmp",
			mockSetup: func(client *MockClient) {
				client.On("GetGroups", mock.Anything, mock.Anything, "master", mock.Anything).
					Return(nil, errBoom)
			},
			expectedError: errBoom,
		},
		{
			name: "Should stop and propagate an error creating the top-level group",
			path: "cmp",
			mockSetup: func(client *MockClient) {
				client.On("GetGroups", mock.Anything, mock.Anything, "master", mock.Anything).
					Return([]*gocloak.Group{}, nil)
				client.On("CreateGroup", mock.Anything, mock.Anything, "master", mock.Anything).
					Return("", errBoom)
			},
			expectedError: errBoom,
		},
		{
			name: "Should stop and propagate an error creating a non-first segment",
			path: "cmp/PROJ001",
			mockSetup: func(client *MockClient) {
				client.On("GetGroups", mock.Anything, mock.Anything, "master", mock.Anything).
					Return([]*gocloak.Group{groupAt("cmp-id", "cmp", "/cmp")}, nil)
				client.On("GetGroup", mock.Anything, mock.Anything, "master", "cmp-id").
					Return(&gocloak.Group{ID: strPtr("cmp-id"), Name: strPtr("cmp"), Path: strPtr("/cmp"), SubGroups: &[]gocloak.Group{}}, nil)
				client.On("CreateChildGroup", mock.Anything, mock.Anything, "master", "cmp-id", mock.Anything).
					Return("", errBoom)
			},
			expectedError: errBoom,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := NewMockClient(t)
			tc.mockSetup(client)
			h := &Helper{Client: client}

			got, err := h.GetOrCreateGroupPath("master", tc.path)

			require.ErrorIs(t, err, tc.expectedError)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestHelperFindChildGroup_TopLevel(t *testing.T) {
	tests := []struct {
		name     string
		groups   []*gocloak.Group
		expected *gocloak.Group
	}{
		{
			name:     "Should find a top-level group by name",
			groups:   []*gocloak.Group{groupNamed("other-id", "other"), groupNamed("cmp-id", "cmp")},
			expected: groupNamed("cmp-id", "cmp"),
		},
		{
			name:     "Should return nil when no top-level group matches",
			groups:   []*gocloak.Group{groupNamed("other-id", "other")},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := NewMockClient(t)
			client.On("GetGroups", mock.Anything, mock.Anything, "master", mock.Anything).Return(tc.groups, nil)

			h := &Helper{Client: client}
			got, err := h.findChildGroup("master", nil, "cmp")

			require.NoError(t, err)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestHelperFindChildGroup_TopLevel_Error(t *testing.T) {
	errBoom := errors.New("boom")
	client := NewMockClient(t)
	client.On("GetGroups", mock.Anything, mock.Anything, "master", mock.Anything).Return(nil, errBoom)

	h := &Helper{Client: client}
	_, err := h.findChildGroup("master", nil, "cmp")

	require.ErrorIs(t, err, errBoom)
}

func TestHelperFindChildGroup_Nested(t *testing.T) {
	parent := groupNamed("parent-id", "parent")
	errBoom := errors.New("boom")

	tests := []struct {
		name          string
		parentGroup   *gocloak.Group
		fetchErr      error
		expected      *gocloak.Group
		expectedError error
	}{
		{
			name:        "Should return nil when the parent has no sub groups",
			parentGroup: &gocloak.Group{ID: parent.ID, Name: parent.Name, SubGroups: nil},
			expected:    nil,
		},
		{
			name: "Should find a matching sub group under the parent",
			parentGroup: &gocloak.Group{
				ID: parent.ID, Name: parent.Name,
				SubGroups: &[]gocloak.Group{*groupNamed("other-id", "other"), *groupNamed("target-id", "target")},
			},
			expected: groupNamed("target-id", "target"),
		},
		{
			name: "Should return nil when no sub group matches",
			parentGroup: &gocloak.Group{
				ID: parent.ID, Name: parent.Name,
				SubGroups: &[]gocloak.Group{*groupNamed("other-id", "other")},
			},
			expected: nil,
		},
		{
			name:          "Should propagate an error fetching the parent group",
			fetchErr:      errBoom,
			expectedError: errBoom,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := NewMockClient(t)
			client.On("GetGroup", mock.Anything, mock.Anything, "master", "parent-id").
				Return(tc.parentGroup, tc.fetchErr)

			h := &Helper{Client: client}
			got, err := h.findChildGroup("master", parent, "target")

			require.ErrorIs(t, err, tc.expectedError)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestHelperCreateChildGroup(t *testing.T) {
	errBoom := errors.New("boom")

	t.Run("Should create a top-level group with Path set to /name", func(t *testing.T) {
		client := NewMockClient(t)
		client.On("CreateGroup", mock.Anything, mock.Anything, "master", mock.Anything).Return("new-id", nil)

		h := &Helper{Client: client}
		got, err := h.createChildGroup("master", nil, "cmp")

		require.NoError(t, err)
		require.Equal(t, groupAt("new-id", "cmp", "/cmp"), got)
	})

	t.Run("Should create a child group with Path derived from the parent's Path", func(t *testing.T) {
		client := NewMockClient(t)
		client.On("CreateChildGroup", mock.Anything, mock.Anything, "master", "parent-id", mock.Anything).Return("child-id", nil)

		h := &Helper{Client: client}
		got, err := h.createChildGroup("master", groupAt("parent-id", "parent", "/parent"), "admin")

		require.NoError(t, err)
		require.Equal(t, groupAt("child-id", "admin", "/parent/admin"), got)
	})

	t.Run("Should propagate an error creating a top-level group", func(t *testing.T) {
		client := NewMockClient(t)
		client.On("CreateGroup", mock.Anything, mock.Anything, "master", mock.Anything).Return("", errBoom)

		h := &Helper{Client: client}
		_, err := h.createChildGroup("master", nil, "cmp")

		require.ErrorIs(t, err, errBoom)
	})

	t.Run("Should propagate an error creating a child group", func(t *testing.T) {
		client := NewMockClient(t)
		client.On("CreateChildGroup", mock.Anything, mock.Anything, "master", "parent-id", mock.Anything).Return("", errBoom)

		h := &Helper{Client: client}
		_, err := h.createChildGroup("master", groupAt("parent-id", "parent", "/parent"), "admin")

		require.ErrorIs(t, err, errBoom)
	})
}

func TestPaginateGroups(t *testing.T) {
	t.Run("Should fetch a single page as-is when the caller already set Max", func(t *testing.T) {
		calls := 0
		fetch := func(p gocloak.GetGroupsParams) ([]*gocloak.Group, error) {
			calls++
			return []*gocloak.Group{groupNamed("a", "a")}, nil
		}

		got, err := paginateGroups(fetch, gocloak.GetGroupsParams{Max: gocloak.IntP(10)})

		require.NoError(t, err)
		require.Equal(t, 1, calls)
		require.Len(t, got, 1)
	})

	t.Run("Should keep paging while pages come back full, and stop on a short page", func(t *testing.T) {
		var seenFirsts []int
		fetch := func(p gocloak.GetGroupsParams) ([]*gocloak.Group, error) {
			seenFirsts = append(seenFirsts, *p.First)
			if *p.First == 0 {
				return make([]*gocloak.Group, groupsPageSize), nil
			}
			return []*gocloak.Group{groupNamed("last", "last")}, nil
		}

		got, err := paginateGroups(fetch, gocloak.GetGroupsParams{})

		require.NoError(t, err)
		require.Equal(t, []int{0, groupsPageSize}, seenFirsts)
		require.Len(t, got, groupsPageSize+1)
	})

	t.Run("Should propagate an error from any page", func(t *testing.T) {
		errBoom := errors.New("boom")
		fetch := func(p gocloak.GetGroupsParams) ([]*gocloak.Group, error) {
			return nil, errBoom
		}

		_, err := paginateGroups(fetch, gocloak.GetGroupsParams{})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestHelperGetUserGroups(t *testing.T) {
	errBoom := errors.New("boom")

	tests := []struct {
		name          string
		groups        []*gocloak.Group
		fetchError    error
		expectedError error
	}{
		{
			name:          "Should propagate an error fetching the user's groups",
			fetchError:    errBoom,
			expectedError: errBoom,
		},
		{
			name:   "Should return the user's groups",
			groups: []*gocloak.Group{groupNamed("g1", "g1")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := NewMockClient(t)
			client.On("GetUserGroups", mock.Anything, mock.Anything, "master", "user-id", mock.Anything).
				Return(tc.groups, tc.fetchError)

			h := &Helper{Client: client}
			got, err := h.GetUserGroups("master", "user-id")

			require.ErrorIs(t, err, tc.expectedError)
			require.Equal(t, tc.groups, got)
		})
	}
}

func TestPaginateGroupMembers(t *testing.T) {
	t.Run("Should fetch a single page as-is when the caller already set Max", func(t *testing.T) {
		calls := 0
		fetch := func(p gocloak.GetGroupsParams) ([]*gocloak.User, error) {
			calls++
			return []*gocloak.User{{Username: strPtr("alice")}}, nil
		}

		got, err := paginateGroupMembers(fetch, gocloak.GetGroupsParams{Max: gocloak.IntP(10)})

		require.NoError(t, err)
		require.Equal(t, 1, calls)
		require.Len(t, got, 1)
	})

	t.Run("Should keep paging while pages come back full, and stop on a short page", func(t *testing.T) {
		var seenFirsts []int
		fetch := func(p gocloak.GetGroupsParams) ([]*gocloak.User, error) {
			seenFirsts = append(seenFirsts, *p.First)
			if *p.First == 0 {
				return make([]*gocloak.User, groupsPageSize), nil
			}
			return []*gocloak.User{{Username: strPtr("last")}}, nil
		}

		got, err := paginateGroupMembers(fetch, gocloak.GetGroupsParams{})

		require.NoError(t, err)
		require.Equal(t, []int{0, groupsPageSize}, seenFirsts)
		require.Len(t, got, groupsPageSize+1)
	})

	t.Run("Should propagate an error from any page", func(t *testing.T) {
		errBoom := errors.New("boom")
		fetch := func(p gocloak.GetGroupsParams) ([]*gocloak.User, error) {
			return nil, errBoom
		}

		_, err := paginateGroupMembers(fetch, gocloak.GetGroupsParams{})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestHelperGetGroupMembers(t *testing.T) {
	errBoom := errors.New("boom")

	tests := []struct {
		name          string
		members       []*gocloak.User
		fetchError    error
		expectedError error
	}{
		{
			name:          "Should propagate an error fetching the group's members",
			fetchError:    errBoom,
			expectedError: errBoom,
		},
		{
			name:    "Should return the group's members",
			members: []*gocloak.User{{Username: strPtr("alice")}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := NewMockClient(t)
			client.On("GetGroupMembers", mock.Anything, mock.Anything, "master", "group-id", mock.Anything).
				Return(tc.members, tc.fetchError)

			h := &Helper{Client: client}
			got, err := h.GetGroupMembers("master", "group-id")

			require.ErrorIs(t, err, tc.expectedError)
			require.Equal(t, tc.members, got)
		})
	}
}

func TestHelperCheckLoginUser(t *testing.T) {
	errBoom := errors.New("boom")

	tests := []struct {
		name                  string
		client                ClientCredentials
		tlsInsecureSkipVerify bool
		token                 *gocloak.JWT
		loginError            error
		expected              *gocloak.JWT
		expectedError         error
	}{
		{
			name:                  "Should return the JWT on a successful login with the default admin client",
			client:                DefaultAdmin,
			tlsInsecureSkipVerify: false,
			token:                 &gocloak.JWT{AccessToken: "access-token"},
			loginError:            nil,
			expected:              &gocloak.JWT{AccessToken: "access-token"},
			expectedError:         nil,
		},
		{
			name:                  "Should authenticate against a caller-supplied client instead of the default",
			client:                ClientCredentials{ID: "custom-client", Secret: "custom-secret"},
			tlsInsecureSkipVerify: false,
			token:                 &gocloak.JWT{AccessToken: "access-token"},
			loginError:            nil,
			expected:              &gocloak.JWT{AccessToken: "access-token"},
			expectedError:         nil,
		},
		{
			name:                  "Should set an insecure TLS config before logging in when configured",
			client:                DefaultAdmin,
			tlsInsecureSkipVerify: true,
			token:                 &gocloak.JWT{AccessToken: "access-token"},
			loginError:            nil,
			expected:              &gocloak.JWT{AccessToken: "access-token"},
			expectedError:         nil,
		},
		{
			name:                  "Should propagate an error when the login fails",
			client:                DefaultAdmin,
			tlsInsecureSkipVerify: false,
			token:                 nil,
			loginError:            errBoom,
			expected:              nil,
			expectedError:         errBoom,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := NewMockClient(t)
			if tc.tlsInsecureSkipVerify {
				client.On("RestyClient").Return(resty.New())
			}
			client.On("Login", mock.Anything, tc.client.ID, tc.client.Secret, "master", "user", "pass").
				Return(tc.token, tc.loginError)

			h := &Helper{
				Client:  client,
				Options: Options{Auth: Auth{Realm: "master"}, Host: Host{TlsInsecureSkipVerify: tc.tlsInsecureSkipVerify}},
			}
			got, err := h.CheckLoginUser("user", "pass", tc.client)

			require.ErrorIs(t, err, tc.expectedError)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestDefaultAdmin(t *testing.T) {
	require.Equal(t, ClientCredentials{ID: "admin-cli", Secret: ""}, DefaultAdmin)
}

func TestHasClientRole(t *testing.T) {
	errBoom := errors.New("boom")

	tests := []struct {
		name          string
		roles         []*gocloak.Role
		fetchError    error
		expected      bool
		expectedError error
	}{
		{
			name:          "Should propagate an error fetching the user's client roles",
			fetchError:    errBoom,
			expectedError: errBoom,
		},
		{
			name: "Should return true when the role is present",
			roles: []*gocloak.Role{
				{Name: strPtr("some-other-role")},
				{Name: strPtr("super-admin")},
			},
			expected: true,
		},
		{
			name:     "Should return false when the role is absent",
			roles:    []*gocloak.Role{{Name: strPtr("some-other-role")}},
			expected: false,
		},
		{
			name:     "Should return false without panicking when a role has a nil name",
			roles:    []*gocloak.Role{{Name: nil}},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := NewMockClient(t)
			client.On("GetClientRolesByUserID", mock.Anything, mock.Anything, "master", "client-id", "user-id").
				Return(tc.roles, tc.fetchError)

			h := &Helper{Client: client}
			got, err := h.HasClientRole("master", "client-id", "user-id", "super-admin")

			require.ErrorIs(t, err, tc.expectedError)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestGetUser(t *testing.T) {
	errBoom := errors.New("boom")

	tests := []struct {
		name          string
		users         []*gocloak.User
		fetchError    error
		expectedUser  *gocloak.User
		expectedError error
	}{
		{
			name:          "Should propagate an error fetching users",
			fetchError:    errBoom,
			expectedError: errBoom,
		},
		{
			name:          "Should return ErrUserNotFound when no user matches",
			users:         []*gocloak.User{},
			expectedError: ErrUserNotFound,
		},
		{
			name:         "Should return the matching user",
			users:        []*gocloak.User{{Username: strPtr("alice")}},
			expectedUser: &gocloak.User{Username: strPtr("alice")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := NewMockClient(t)
			client.On("GetUsers", mock.Anything, mock.Anything, "master", gocloak.GetUsersParams{
				Username: strPtr("alice"),
				Exact:    gocloak.BoolP(true),
			}).Return(tc.users, tc.fetchError)

			h := &Helper{Client: client}
			got, err := h.GetUser("master", "alice")

			require.ErrorIs(t, err, tc.expectedError)
			require.Equal(t, tc.expectedUser, got)
		})
	}
}

func TestLoginServiceAccount(t *testing.T) {
	errBoom := errors.New("boom")

	tests := []struct {
		name          string
		token         *gocloak.JWT
		loginError    error
		expectedToken string
		expectedError string
	}{
		{
			name:          "Should set the Helper's Token from the returned JWT on success",
			token:         &gocloak.JWT{AccessToken: "fresh-token"},
			expectedToken: "fresh-token",
		},
		{
			name:          "Should return a wrapped error and leave Token unset if the login fails",
			loginError:    errBoom,
			expectedToken: "",
			expectedError: "keycloak service account login failed: boom",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := NewMockClient(t)
			client.On("LoginClient", mock.Anything, "cmp-migration-sa", "sa-secret", "master").
				Return(tc.token, tc.loginError)

			h := &Helper{
				Client: client,
				Options: Options{
					Auth: Auth{
						Realm:        "master",
						ClientID:     "cmp-migration-sa",
						ClientSecret: "sa-secret",
					},
				},
			}
			err := h.LoginServiceAccount()

			if tc.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedError)
			}
			require.Equal(t, tc.expectedToken, h.Token)
		})
	}
}

func TestSetKeycloakClient(t *testing.T) {
	tests := []struct {
		name          string
		options       Options
		expectedError string
	}{
		{
			name:          "Should return an error if scheme is empty",
			options:       Options{Host: Host{Ip: "keycloak", Port: 80, Path: "auth"}, Auth: Auth{Realm: "master", Username: "admin", Password: "admin"}},
			expectedError: "keycloak scheme is empty",
		},
		{
			name:          "Should return an error if ip is empty",
			options:       Options{Host: Host{Scheme: "http", Port: 80, Path: "auth"}, Auth: Auth{Realm: "master", Username: "admin", Password: "admin"}},
			expectedError: "keycloak ip is empty",
		},
		{
			name:          "Should return an error if port is empty",
			options:       Options{Host: Host{Scheme: "http", Ip: "keycloak", Path: "auth"}, Auth: Auth{Realm: "master", Username: "admin", Password: "admin"}},
			expectedError: "keycloak port is empty",
		},
		{
			name:          "Should return an error if path is empty",
			options:       Options{Host: Host{Scheme: "http", Ip: "keycloak", Port: 80}, Auth: Auth{Realm: "master", Username: "admin", Password: "admin"}},
			expectedError: "keycloak path is empty",
		},
		{
			name:          "Should return an error if neither username/password nor clientId/clientSecret is set",
			options:       Options{Host: Host{Scheme: "http", Ip: "keycloak", Port: 80, Path: "auth"}, Auth: Auth{Realm: "master"}},
			expectedError: "keycloak credentials are empty: need either username/password or clientId/clientSecret",
		},
		{
			name:          "Should return an error if only username is set, without a password",
			options:       Options{Host: Host{Scheme: "http", Ip: "keycloak", Port: 80, Path: "auth"}, Auth: Auth{Realm: "master", Username: "admin"}},
			expectedError: "keycloak credentials are empty: need either username/password or clientId/clientSecret",
		},
		{
			name:          "Should return an error if only clientId is set, without a clientSecret",
			options:       Options{Host: Host{Scheme: "http", Ip: "keycloak", Port: 80, Path: "auth"}, Auth: Auth{Realm: "master", ClientID: "cmp-migration-sa"}},
			expectedError: "keycloak credentials are empty: need either username/password or clientId/clientSecret",
		},
		{
			name:          "Should return an error if realm is empty",
			options:       Options{Host: Host{Scheme: "http", Ip: "keycloak", Port: 80, Path: "auth"}, Auth: Auth{Username: "admin", Password: "admin"}},
			expectedError: "keycloak realm is empty",
		},
		{
			name:    "Should succeed with username/password credentials",
			options: Options{Host: Host{Scheme: "http", Ip: "keycloak", Port: 80, Path: "auth"}, Auth: Auth{Realm: "master", Username: "admin", Password: "admin"}},
		},
		{
			name:    "Should succeed with clientId/clientSecret credentials, without any username/password set",
			options: Options{Host: Host{Scheme: "http", Ip: "keycloak", Port: 80, Path: "auth"}, Auth: Auth{Realm: "master", ClientID: "cmp-migration-sa", ClientSecret: "sa-secret"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := &Helper{Options: tc.options}
			err := h.SetKeycloakClient()

			if tc.expectedError == "" {
				require.NoError(t, err)
				require.NotNil(t, h.Client)
			} else {
				require.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestHelperExecuteActionsEmail(t *testing.T) {
	errBoom := errors.New("boom")

	tests := []struct {
		name            string
		actions         []string
		expectedActions []string
		clientError     error
		expectedError   error
	}{
		{
			name:            "Should send the given actions when actions is non-nil",
			actions:         []string{"UPDATE_PASSWORD"},
			expectedActions: []string{"UPDATE_PASSWORD"},
			clientError:     nil,
			expectedError:   nil,
		},
		{
			name:            "Should send an empty slice instead of nil when actions is nil",
			actions:         nil,
			expectedActions: []string{},
			clientError:     nil,
			expectedError:   nil,
		},
		{
			name:            "Should propagate an error when the client call fails",
			actions:         []string{"UPDATE_PASSWORD"},
			expectedActions: []string{"UPDATE_PASSWORD"},
			clientError:     errBoom,
			expectedError:   errBoom,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := NewMockClient(t)
			client.On("ExecuteActionsEmail", mock.Anything, mock.Anything, "master", mock.Anything).
				Run(func(args mock.Arguments) {
					params := args.Get(3).(gocloak.ExecuteActionsEmail)
					require.NotNil(t, params.UserID)
					require.Equal(t, "user-id", *params.UserID)
					require.NotNil(t, params.Actions)
					require.Equal(t, tc.expectedActions, *params.Actions)
				}).
				Return(tc.clientError)

			h := &Helper{Client: client}
			err := h.ExecuteActionsEmail("master", "user-id", tc.actions)

			require.ErrorIs(t, err, tc.expectedError)
		})
	}
}
