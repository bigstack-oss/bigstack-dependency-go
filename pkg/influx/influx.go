package influx

import (
	"context"
	"sync"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	influxv2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

var (
	helper *Helper
	once   sync.Once
)

type ConnectClient interface {
	QueryAPI(string) api.QueryAPI
	Close()
}

type QueryApiClient interface {
	Query(context.Context, string) (*api.QueryTableResult, error)
}

type Helper struct {
	ConnectClient
	QueryApiClient
	Options
}

func initOptions(opts []Option) *Options {
	options := &Options{Auth: Auth{}}
	for _, o := range opts {
		o(options)
	}

	return options
}

func NewHelper(opts ...Option) (*Helper, error) {
	initedOpts := initOptions(opts)

	h := &Helper{Options: *initedOpts}
	h.ConnectClient = influxv2.NewClient(h.Options.Url, h.Options.Auth.Token)
	h.QueryApiClient = h.ConnectClient.QueryAPI(h.Options.Org)

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

func GetQueryCursor(stmt string) (*api.QueryTableResult, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(60))
	c, err := helper.QueryApiClient.Query(ctx, stmt)
	return c, cancel, err
}

func (h *Helper) Close() {
	if h.ConnectClient == nil {
		return
	}

	h.ConnectClient.Close()
}
