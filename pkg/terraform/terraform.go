package terraform

import (
	"context"
	"fmt"
	"sync"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	CreateRecordIfNotExist = options.Update().SetUpsert(true)

	helper *Helper
	once   sync.Once
)

type Client interface {
	Show(context.Context, ...tfexec.ShowOption) (*tfjson.State, error)
}

type Helper struct {
	Client
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

	err := h.SetTerraformClient()
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

func (h *Helper) SetTerraformClient() error {
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion(h.Options.Version)),
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(120))
	defer cancel()
	execPath, err := installer.Install(ctx)
	if err != nil {
		return err
	}

	h.Client, err = tfexec.NewTerraform(h.Options.WoringDir, execPath)
	if err != nil {
		return err
	}

	return nil
}

func GetGlobalHelper() *Helper {
	return helper
}

func (h *Helper) ShowResourceValues(resourceType string) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(120))
	defer cancel()

	state, err := h.Show(ctx)
	if err != nil {
		return nil, err
	}

	for _, child := range state.Values.RootModule.ChildModules {
		for _, resource := range child.Resources {
			if resource.Type == resourceType {
				return resource.AttributeValues, nil
			}
		}
	}

	return nil, fmt.Errorf(
		"resource type %s not found in state",
		resourceType,
	)
}
