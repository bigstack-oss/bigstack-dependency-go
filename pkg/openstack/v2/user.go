package openstack

import (
	"context"
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/users"
)

func (h *Helper) ListUsers(opts *users.ListOpts) ([]users.User, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := users.List(h.Identity, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	users, err := users.ExtractUsers(pages)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (h *Helper) GetUserIdByName(name string) (string, error) {
	users, err := h.ListUsers(&users.ListOpts{Name: name})
	if err != nil {
		return "", err
	}

	userId := ""
	for _, user := range users {
		if user.Name == name {
			userId = user.ID
			break
		}
	}
	if userId == "" {
		return "", fmt.Errorf("user %s not found", name)
	}

	return userId, nil
}

func (h *Helper) CreateUser(opts users.CreateOpts) (*users.User, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return users.Create(
		ctx,
		h.Identity,
		opts,
	).Extract()
}

func (h *Helper) GetUserByName(name string) (*users.User, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	pages, err := users.List(h.Identity, &users.ListOpts{Name: name}).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	users, err := users.ExtractUsers(pages)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Name == name {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("user %s not found", name)
}
