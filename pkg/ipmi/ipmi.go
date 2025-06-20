package ipmi

import (
	"context"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bougou/go-ipmi"
)

type Client interface {
	Connect(context.Context) error
	GetFRU(context.Context, uint8, string) (*ipmi.FRU, error)
	Close(context.Context) error
}

type Helper struct {
	cancel *context.CancelFunc
	Client
	Options
}

func initOptions(opts []Option) *Options {
	options := genDefaultOptions()
	for _, o := range opts {
		o(options)
	}

	return options
}

func genDefaultOptions() *Options {
	return &Options{Port: defaultPort}
}

func NewHelper(opts ...Option) (*Helper, error) {
	initedOpts := initOptions(opts)
	client, err := ipmi.NewClient(
		initedOpts.Host,
		initedOpts.Port,
		initedOpts.Username,
		initedOpts.Password,
	)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(20))
	err = client.Connect(ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	return &Helper{
		Client:  client,
		Options: *initedOpts,
		cancel:  &cancel,
	}, nil
}

func (h *Helper) GetFRU(deviceId uint8) (*ipmi.FRU, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.GetFRU(
		ctx,
		deviceId,
		"",
	)
}

func (h *Helper) Close() error {
	if h.Client == nil {
		return nil
	}

	if h.cancel != nil {
		(*h.cancel)()
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	return h.Client.Close(ctx)
}
