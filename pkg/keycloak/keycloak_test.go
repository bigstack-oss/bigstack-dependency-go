package keycloak

import (
	"errors"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	keycloakMocks "github.com/bigstack-oss/bigstack-dependency-go/pkg/keycloak/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }

func groupNamed(id, name string) *gocloak.Group {
	return &gocloak.Group{ID: strPtr(id), Name: strPtr(name)}
}

func TestEnsureGroupPath_InvalidPath(t *testing.T) {
	h := &Helper{Client: keycloakMocks.NewMockClient(t)}
	_, err := h.EnsureGroupPath("master", "")
	require.EqualError(t, err, `invalid group path ""`)
}

func TestEnsureGroupPath(t *testing.T) {
	errBoom := errors.New("boom")

	tests := []struct {
		name          string
		path          string
		mockSetup     func(client *keycloakMocks.MockClient)
		expected      *gocloak.Group
		expectedError error
	}{
		{
			name: "Should return the existing group for a single-segment path found at the top level",
			path: "cmp",
			mockSetup: func(client *keycloakMocks.MockClient) {
				client.On("GetGroups", mock.Anything, mock.Anything, "master", mock.Anything).
					Return([]*gocloak.Group{groupNamed("cmp-id", "cmp")}, nil)
			},
			expected: groupNamed("cmp-id", "cmp"),
		},
		{
			name: "Should create a single-segment path not found at the top level",
			path: "cmp",
			mockSetup: func(client *keycloakMocks.MockClient) {
				client.On("GetGroups", mock.Anything, mock.Anything, "master", mock.Anything).
					Return([]*gocloak.Group{}, nil)
				client.On("CreateGroup", mock.Anything, mock.Anything, "master", mock.Anything).
					Return("cmp-id", nil)
			},
			expected: groupNamed("cmp-id", "cmp"),
		},
		{
			name: "Should walk every segment of a multi-segment path that already exists",
			path: "cmp/PROJ001/admin",
			mockSetup: func(client *keycloakMocks.MockClient) {
				client.On("GetGroups", mock.Anything, mock.Anything, "master", mock.Anything).
					Return([]*gocloak.Group{groupNamed("cmp-id", "cmp")}, nil)
				client.On("GetGroup", mock.Anything, mock.Anything, "master", "cmp-id").
					Return(&gocloak.Group{
						ID: strPtr("cmp-id"), Name: strPtr("cmp"),
						SubGroups: &[]gocloak.Group{*groupNamed("proj-id", "PROJ001")},
					}, nil)
				client.On("GetGroup", mock.Anything, mock.Anything, "master", "proj-id").
					Return(&gocloak.Group{
						ID: strPtr("proj-id"), Name: strPtr("PROJ001"),
						SubGroups: &[]gocloak.Group{*groupNamed("admin-id", "admin")},
					}, nil)
			},
			expected: groupNamed("admin-id", "admin"),
		},
		{
			name: "Should create only the missing middle segment of a multi-segment path",
			path: "cmp/PROJ001/admin",
			mockSetup: func(client *keycloakMocks.MockClient) {
				client.On("GetGroups", mock.Anything, mock.Anything, "master", mock.Anything).
					Return([]*gocloak.Group{groupNamed("cmp-id", "cmp")}, nil)
				client.On("GetGroup", mock.Anything, mock.Anything, "master", "cmp-id").
					Return(&gocloak.Group{ID: strPtr("cmp-id"), Name: strPtr("cmp"), SubGroups: &[]gocloak.Group{}}, nil)
				client.On("CreateChildGroup", mock.Anything, mock.Anything, "master", "cmp-id", mock.Anything).
					Return("proj-id", nil)
				client.On("GetGroup", mock.Anything, mock.Anything, "master", "proj-id").
					Return(&gocloak.Group{
						ID: strPtr("proj-id"), Name: strPtr("PROJ001"),
						SubGroups: &[]gocloak.Group{*groupNamed("admin-id", "admin")},
					}, nil)
			},
			expected: groupNamed("admin-id", "admin"),
		},
		{
			name: "Should stop and propagate an error finding the top-level group",
			path: "cmp",
			mockSetup: func(client *keycloakMocks.MockClient) {
				client.On("GetGroups", mock.Anything, mock.Anything, "master", mock.Anything).
					Return(nil, errBoom)
			},
			expectedError: errBoom,
		},
		{
			name: "Should stop and propagate an error creating the top-level group",
			path: "cmp",
			mockSetup: func(client *keycloakMocks.MockClient) {
				client.On("GetGroups", mock.Anything, mock.Anything, "master", mock.Anything).
					Return([]*gocloak.Group{}, nil)
				client.On("CreateGroup", mock.Anything, mock.Anything, "master", mock.Anything).
					Return("", errBoom)
			},
			expectedError: errBoom,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := keycloakMocks.NewMockClient(t)
			tc.mockSetup(client)
			h := &Helper{Client: client}

			got, err := h.EnsureGroupPath("master", tc.path)

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
			client := keycloakMocks.NewMockClient(t)
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
	client := keycloakMocks.NewMockClient(t)
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
			client := keycloakMocks.NewMockClient(t)
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

	t.Run("Should create a top-level group when parent is nil", func(t *testing.T) {
		client := keycloakMocks.NewMockClient(t)
		client.On("CreateGroup", mock.Anything, mock.Anything, "master", mock.Anything).Return("new-id", nil)

		h := &Helper{Client: client}
		got, err := h.createChildGroup("master", nil, "cmp")

		require.NoError(t, err)
		require.Equal(t, "new-id", *got.ID)
		require.Equal(t, "cmp", *got.Name)
	})

	t.Run("Should create a child group under the parent", func(t *testing.T) {
		client := keycloakMocks.NewMockClient(t)
		client.On("CreateChildGroup", mock.Anything, mock.Anything, "master", "parent-id", mock.Anything).Return("child-id", nil)

		h := &Helper{Client: client}
		got, err := h.createChildGroup("master", groupNamed("parent-id", "parent"), "admin")

		require.NoError(t, err)
		require.Equal(t, "child-id", *got.ID)
	})

	t.Run("Should propagate an error creating a top-level group", func(t *testing.T) {
		client := keycloakMocks.NewMockClient(t)
		client.On("CreateGroup", mock.Anything, mock.Anything, "master", mock.Anything).Return("", errBoom)

		h := &Helper{Client: client}
		_, err := h.createChildGroup("master", nil, "cmp")

		require.ErrorIs(t, err, errBoom)
	})

	t.Run("Should propagate an error creating a child group", func(t *testing.T) {
		client := keycloakMocks.NewMockClient(t)
		client.On("CreateChildGroup", mock.Anything, mock.Anything, "master", "parent-id", mock.Anything).Return("", errBoom)

		h := &Helper{Client: client}
		_, err := h.createChildGroup("master", groupNamed("parent-id", "parent"), "admin")

		require.ErrorIs(t, err, errBoom)
	})
}

func TestCmpProjectRoleFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     *string
		expected CmpProjectRole
		expectOk bool
	}{
		{
			name:     "Should parse a three-segment path",
			path:     strPtr("/cmp/PROJ001/admin"),
			expected: CmpProjectRole{Product: "cmp", Project: "PROJ001", Role: "admin"},
			expectOk: true,
		},
		{
			name:     "Should return not-ok for a nil path",
			path:     nil,
			expectOk: false,
		},
		{
			name:     "Should return not-ok for a one-segment path",
			path:     strPtr("/cmp"),
			expectOk: false,
		},
		{
			name:     "Should return not-ok for a two-segment path",
			path:     strPtr("/cmp/PROJ001"),
			expectOk: false,
		},
		{
			name:     "Should return not-ok for a path with more than three segments",
			path:     strPtr("/cmp/PROJ001/admin/extra"),
			expectOk: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := CmpProjectRoleFromPath(tc.path)

			require.Equal(t, tc.expectOk, ok)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestCmpProjectRolesFromGroups(t *testing.T) {
	tests := []struct {
		name     string
		groups   []*gocloak.Group
		expected []CmpProjectRole
	}{
		{
			name: "Should parse multiple project/role memberships",
			groups: []*gocloak.Group{
				{Path: strPtr("/cmp/PROJ001/admin")},
				{Path: strPtr("/cmp/PROJ002/member")},
			},
			expected: []CmpProjectRole{
				{Product: "cmp", Project: "PROJ001", Role: "admin"},
				{Product: "cmp", Project: "PROJ002", Role: "member"},
			},
		},
		{
			name: "Should skip groups with a nil or non-three-segment path",
			groups: []*gocloak.Group{
				{Path: nil},
				{Path: strPtr("/cmp")},
				{Path: strPtr("/cmp/PROJ001/admin")},
			},
			expected: []CmpProjectRole{{Product: "cmp", Project: "PROJ001", Role: "admin"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, CmpProjectRolesFromGroups(tc.groups))
		})
	}
}

func TestGetCmpProjectRoles(t *testing.T) {
	errBoom := errors.New("boom")

	tests := []struct {
		name          string
		groups        []*gocloak.Group
		fetchError    error
		expected      []CmpProjectRole
		expectedError error
	}{
		{
			name:          "Should propagate an error fetching the user's groups",
			fetchError:    errBoom,
			expectedError: errBoom,
		},
		{
			name:     "Should delegate parsing to CmpProjectRolesFromGroups on success",
			groups:   []*gocloak.Group{{Path: strPtr("/cmp/PROJ001/admin")}},
			expected: []CmpProjectRole{{Product: "cmp", Project: "PROJ001", Role: "admin"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := keycloakMocks.NewMockClient(t)
			client.On("GetUserGroups", mock.Anything, mock.Anything, "master", "user-id", mock.Anything).
				Return(tc.groups, tc.fetchError)

			h := &Helper{Client: client}
			got, err := h.GetCmpProjectRoles("master", "user-id")

			require.ErrorIs(t, err, tc.expectedError)
			require.Equal(t, tc.expected, got)
		})
	}
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
			client := keycloakMocks.NewMockClient(t)
			client.On("GetClientRolesByUserID", mock.Anything, mock.Anything, "master", "client-id", "user-id").
				Return(tc.roles, tc.fetchError)

			h := &Helper{Client: client}
			got, err := h.HasClientRole("master", "client-id", "user-id", "super-admin")

			require.ErrorIs(t, err, tc.expectedError)
			require.Equal(t, tc.expected, got)
		})
	}
}
