// Package email (v2) sends plaintext SMTP email with tunable transport security
// and optional authentication. It is a ground-up successor to the origin
// pkg/email (which wraps jordan-wright/email + net/smtp); v1 stays in place for
// existing consumers (cube-cos-api, quota/billing alerts) while new code opts
// into this package via the .../pkg/email/v2 import path, matching the repo's
// pkg/openstack/v2 idiom.
//
// It wraps github.com/wneessen/go-mail and follows the sibling-package shape
// (Options + functional options, a Client interface the Helper embeds for
// dependency injection, NewHelper -> SetClient, plus a NewGlobalHelper /
// GetGlobalHelper singleton). A Helper holds the reusable SMTP connection; the
// per-message content is passed to Send, so one Helper sends many messages.
// Toggling TLS and Auth lets a caller target a relay with no STARTTLS and no
// auth as easily as a secured one.
package email

import (
	"fmt"
	"sync"
	"time"

	mail "github.com/wneessen/go-mail"
)

var (
	helper *Helper
	once   sync.Once
)

// Client is the SMTP transport the Helper depends on. It is satisfied by
// *mail.Client (wneessen/go-mail) and can be swapped for a fake in tests.
type Client interface {
	DialAndSend(messages ...*mail.Msg) error
}

// Helper holds a reusable SMTP connection. Build it once with NewHelper, then
// call Send for each message.
type Helper struct {
	Client
	Options
}

// Message is the per-send content. From is part of the Sender (set once at
// NewHelper time), so a Message only carries what changes between sends.
type Message struct {
	To      []string
	Subject string
	Body    string
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
		Sender: Sender{TLS: TLSMandatory},
		Retry: Retry{
			Count:       3,
			WaitTime:    2 * time.Second,
			MaxWaitTime: 5 * time.Second,
		},
	}
}

// NewHelper builds a Helper from the Sender options and injects the SMTP client.
// It returns an error if the connection options are invalid (bad TLS policy,
// port out of range, ...).
func NewHelper(opts ...Option) (*Helper, error) {
	initedOpts := initOptions(opts)
	h := &Helper{Options: *initedOpts}

	err := h.SetClient()
	if err != nil {
		return nil, err
	}

	return h, nil
}

// NewGlobalHelper builds the package-level Helper once, for callers that want a
// single app-wide sender configured at startup and reused via GetGlobalHelper.
func NewGlobalHelper(opts ...Option) error {
	var err error
	once.Do(func() {
		helper, err = NewHelper(opts...)
	})

	return err
}

// GetGlobalHelper returns the Helper built by NewGlobalHelper (nil if unset).
func GetGlobalHelper() *Helper {
	return helper
}

// SetClient constructs the go-mail client from the Sender options and assigns it
// to h.Client. With Auth=false go-mail keeps its default NoAuth and never
// attempts AUTH; with Auth=true over a TLS-less link it uses PLAIN-NOENC, since
// standard PLAIN refuses to transmit credentials over an unencrypted link.
func (h *Helper) SetClient() error {
	policy, err := h.TLS.policy()
	if err != nil {
		return err
	}

	opts := []mail.Option{
		mail.WithPort(h.Port),
		mail.WithTLSPolicy(policy),
	}
	if h.Auth {
		authType := mail.SMTPAuthPlain
		if policy == mail.NoTLS {
			authType = mail.SMTPAuthPlainNoEnc
		}
		opts = append(opts,
			mail.WithSMTPAuth(authType),
			mail.WithUsername(h.Username),
			mail.WithPassword(h.Password),
		)
	}

	client, err := mail.NewClient(h.Host, opts...)
	if err != nil {
		return fmt.Errorf("email: build client: %w", err)
	}

	h.Client = client
	return nil
}

// Send delivers msg over the Helper's connection, re-dialing up to Retry.Count
// times on a transient failure.
func (h *Helper) Send(msg Message) error {
	m, err := h.buildMessage(msg)
	if err != nil {
		return err
	}

	var lastErr error
	for trial := 0; trial <= h.Count; trial++ {
		lastErr = h.Client.DialAndSend(m)
		if lastErr == nil {
			return nil
		}
	}

	return fmt.Errorf("email: send failed after %d retries: %w", h.Count, lastErr)
}

func (h *Helper) buildMessage(msg Message) (*mail.Msg, error) {
	m := mail.NewMsg()
	if err := m.From(h.From); err != nil {
		return nil, fmt.Errorf("email: set from %q: %w", h.From, err)
	}
	if err := m.To(msg.To...); err != nil {
		return nil, fmt.Errorf("email: set recipients %v: %w", msg.To, err)
	}
	m.Subject(msg.Subject)
	m.SetBodyString(mail.TypeTextPlain, msg.Body)

	return m, nil
}

// policy maps the public TLS value onto go-mail's TLSPolicy. An empty value
// defaults to TLSMandatory so an unset config never silently downgrades.
func (t TLS) policy() (mail.TLSPolicy, error) {
	switch t {
	case TLSNone:
		return mail.NoTLS, nil
	case TLSOpportunistic:
		return mail.TLSOpportunistic, nil
	case TLSMandatory, "":
		return mail.TLSMandatory, nil
	default:
		return 0, fmt.Errorf("email: invalid TLS policy %q (want none|opportunistic|mandatory)", t)
	}
}
