package ssh

import (
	"time"

	"golang.org/x/crypto/ssh"
)

type Option func(*Options)

type Options struct {
	Host            string
	User            string
	HostKeyCallback ssh.HostKeyCallback
	Timeout         time.Duration
}

func User(user string) Option {
	return func(o *Options) {
		o.User = user
	}
}

func HostKeyCallback(callback ssh.HostKeyCallback) Option {
	return func(o *Options) {
		o.HostKeyCallback = callback
	}
}

func Timeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}
