package slack

import (
	"sync"

	"github.com/slack-go/slack"
)

var (
	helper *Helper
	once   sync.Once
)

type Client interface {
	PostMessage(string, ...slack.MsgOption) (string, string, error)
}

type Helper struct {
	Client
	Options
}

func NewHelper(opts ...Option) (*Helper, error) {
	initedOpts := initOptions(opts)
	h := &Helper{Options: *initedOpts}

	err := h.SetClient()
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

func (h *Helper) SetClient() error {
	h.Client = slack.New(h.Token)
	return nil
}

func initOptions(opts []Option) *Options {
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	return options
}
