package email

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/smtp"
	"time"

	"github.com/jordan-wright/email"
	"go-micro.dev/v5/logger"
)

var (
	Send = smtp.SendMail
)

type Client interface {
	Attach(r io.Reader, filename string, c string) (a *email.Attachment, err error)
	SendWithStartTLS(addr string, a smtp.Auth, t *tls.Config) error
}

type Helper struct {
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
	return &Options{
		Retry: Retry{
			Count:       3,
			WaitTime:    2 * time.Second,
			MaxWaitTime: 5 * time.Second,
		},
	}
}

func NewHelper(opts ...Option) *Helper {
	initedOpts := initOptions(opts)
	e := &email.Email{
		From:        initedOpts.Email.From,
		To:          initedOpts.Email.To,
		Subject:     initedOpts.Email.Subject,
		HTML:        initedOpts.Email.HTML,
		Attachments: initedOpts.Email.Attachments,
	}

	return &Helper{
		Client:  e,
		Options: *initedOpts,
	}
}

func (h *Helper) Send() error {
	var err error
	trialCount := 0

	for {
		if trialCount > h.Options.Count {
			logger.Errorf("failed after retried %d times: %w", h.Options.Retry, err)
			return err
		}

		err = h.SendWithStartTLS(
			h.getSmtpAddress(),
			h.getSmtpAuth(),
			h.getTLSConfig(),
		)
		if err != nil {
			trialCount++
			continue
		}

		return nil
	}
}

func (h *Helper) getSmtpAddress() string {
	return fmt.Sprintf("%s:%s", h.Options.Host, h.Options.Port)
}

func (h *Helper) getSmtpAuth() smtp.Auth {
	return smtp.PlainAuth(
		"",
		h.Options.Username,
		h.Options.Password,
		h.Options.Host,
	)
}

func (h *Helper) getTLSConfig() *tls.Config {
	return &tls.Config{ServerName: h.Host}
}
