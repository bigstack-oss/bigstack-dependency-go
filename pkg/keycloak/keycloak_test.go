package keycloak

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Nerzal/gocloak/v13"
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

func TestEnsureGroupPath_InvalidPath(t *testing.T) {
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
			_, err := h.EnsureGroupPath("master", tc.path)
			require.EqualError(t, err, fmt.Sprintf("invalid group path %q", tc.path))
		})
	}
}

func TestEnsureGroupPath(t *testing.T) {
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
