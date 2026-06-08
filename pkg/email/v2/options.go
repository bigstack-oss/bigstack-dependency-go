package email

import "time"

// Option mutates Options and is passed to NewHelper / NewGlobalHelper.
type Option func(*Options)

// Options is the reusable connection configuration for a Helper: the SMTP
// sender (relay + sending identity) and the resend policy. Per-message content
// (recipients, subject, body) is passed to Send, so one Helper — built once —
// can send many messages.
type Options struct {
	Sender
	Retry
}

// Sender is the SMTP connection and sending identity. TLS and Auth make
// TLS-less / unauthenticated relays a first-class configuration rather than a
// separate code path.
type Sender struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	From     string `json:"from" yaml:"from"`
	TLS      TLS    `json:"tls" yaml:"tls"`
	Auth     bool   `json:"auth" yaml:"auth"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

// Retry controls how many times Send re-dials on a transient failure.
type Retry struct {
	Count       int           `json:"count" yaml:"count"`
	WaitTime    time.Duration `json:"waitTime" yaml:"waitTime"`
	MaxWaitTime time.Duration `json:"maxWaitTime" yaml:"maxWaitTime"`
}

// TLS selects the transport-encryption policy for a connection.
type TLS string

const (
	// TLSNone keeps the conversation plaintext even if the server advertises
	// STARTTLS. Use for relays without TLS.
	TLSNone TLS = "none"
	// TLSOpportunistic upgrades to STARTTLS when advertised, else stays plaintext.
	TLSOpportunistic TLS = "opportunistic"
	// TLSMandatory requires a successful STARTTLS upgrade or the send fails.
	TLSMandatory TLS = "mandatory"
)

func SenderHost(host string) Option {
	return func(o *Options) {
		o.Sender.Host = host
	}
}

func SenderPort(port int) Option {
	return func(o *Options) {
		o.Sender.Port = port
	}
}

func SenderFrom(from string) Option {
	return func(o *Options) {
		o.Sender.From = from
	}
}

func SenderTLS(policy TLS) Option {
	return func(o *Options) {
		o.Sender.TLS = policy
	}
}

func SenderAuth(auth bool) Option {
	return func(o *Options) {
		o.Sender.Auth = auth
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

func RetryCount(count int) Option {
	return func(o *Options) {
		o.Retry.Count = count
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
