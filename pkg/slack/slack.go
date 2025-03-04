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
	PostWebhook func(string, *slack.WebhookMessage) error
	Options
}

func NewHelper(opts ...Option) (*Helper, error) {
	initedOpts := initOptions(opts)
	h := &Helper{Options: *initedOpts}

	err := h.SetClients()
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

func initOptions(opts []Option) *Options {
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	return options
}

func (h *Helper) SetClients() error {
	h.Client = slack.New(h.Token)
	h.PostWebhook = slack.PostWebhook
	return nil
}

func (h *Helper) SendTextMsg(channel, msg string) error {
	_, _, err := h.Client.PostMessage(channel, slack.MsgOptionText(msg, false))
	return err
}

func (h *Helper) SendWebhookMsg(url string, msg string) error {
	return h.PostWebhook(
		url,
		&slack.WebhookMessage{
			Text: msg,
		},
	)
}
