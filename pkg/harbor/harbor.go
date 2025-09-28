package harbor

import (
	"context"
	"sync"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/goharbor/go-client/pkg/harbor"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/project"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"
)

var (
	helper *Helper

	once sync.Once
)

type ProjectCli interface {
	CreateProject(ctx context.Context, params *project.CreateProjectParams) (*project.CreateProjectCreated, error)
}

type Helper struct {
	ProjectCli
	*Options
}

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

func initOptions(opts []Option) *Options {
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	return options
}

func NewHelper(opts ...Option) (*Helper, error) {
	initedOpts := initOptions(opts)

	cli, err := harbor.NewClientSet(&harbor.ClientSetConfig{
		URL:      initedOpts.Url,
		Password: initedOpts.Username,
		Username: initedOpts.Password,
		Insecure: initedOpts.InsecureSkipVerify,
	})
	if err != nil {
		return nil, err
	}

	return &Helper{
		ProjectCli: cli.V2().Project,
		Options:    initedOpts,
	}, nil
}

func (h *Helper) CreateProject(name string) (*project.CreateProjectCreated, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	return h.ProjectCli.CreateProject(
		ctx,
		&project.CreateProjectParams{
			Project: &models.ProjectReq{
				ProjectName: name,
				Metadata: &models.ProjectMetadata{
					Public: "true",
				},
			},
		},
	)
}
