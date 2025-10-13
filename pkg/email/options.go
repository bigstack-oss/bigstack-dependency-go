package email

import (
	"time"

	"github.com/jordan-wright/email"
)

var (
	Opts *Options
)

type Option func(*Options)

type Options struct {
	Email
	Sender
	Retry
}

type Email struct {
	From        string
	To          []string
	Subject     string
	HTML        []byte
	Attachments []*email.Attachment
}

type Sender struct {
	Host     string
	Port     string
	Username string
	Password string
}

type Retry struct {
	Count       int
	WaitTime    time.Duration
	MaxWaitTime time.Duration
}

func EmailFrom(from string) Option {
	return func(o *Options) {
		o.Email.From = from
	}
}

func EmailTo(to []string) Option {
	return func(o *Options) {
		o.Email.To = to
	}
}

func EmailSubject(subject string) Option {
	return func(o *Options) {
		o.Email.Subject = subject
	}
}

func EmailHTML(html []byte) Option {
	return func(o *Options) {
		o.Email.HTML = html
	}
}

func EmailAttachments(attachments []*email.Attachment) Option {
	return func(o *Options) {
		o.Email.Attachments = attachments
	}
}

func SenderHost(host string) Option {
	return func(o *Options) {
		o.Sender.Host = host
	}
}

func SenderPort(port string) Option {
	return func(o *Options) {
		o.Sender.Port = port
	}
}

func SenderUsername(username string) Option {
	return func(o *Options) {
		o.Sender.Username = username
	}
}

func SenderPassword(password string) Option {
	return func(o *Options) {
		o.Sender.Password = password
	}
}

func RetryWaitTime(t time.Duration) Option {
	return func(o *Options) {
		o.Retry.WaitTime = t
	}
}

func RetryMaxWaitTime(t time.Duration) Option {
	return func(o *Options) {
		o.Retry.MaxWaitTime = t
	}
}
