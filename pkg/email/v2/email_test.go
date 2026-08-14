package email

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"strings"
	"testing"

	mail "github.com/wneessen/go-mail"
)

// fakeClient is an injected Client (dependency injection) that records the
// messages handed to it instead of dialing a server.
type fakeClient struct {
	msgs []*mail.Msg
	err  error
}

func (f *fakeClient) DialAndSend(messages ...*mail.Msg) error {
	f.msgs = append(f.msgs, messages...)
	return f.err
}

// TestSendDelegatesToInjectedClient proves a Helper holds the connection and
// Send takes the per-message content, delivering it through the Client interface.
func TestSendDelegatesToInjectedClient(t *testing.T) {
	fc := &fakeClient{}
	h := &Helper{
		Client:  fc,
		Options: Options{Sender: Sender{From: "noreply@bigstack.co"}},
	}

	err := h.Send(Message{To: []string{"user@example.com"}, Subject: "hi", Body: "body"})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if len(fc.msgs) != 1 {
		t.Fatalf("expected 1 message delivered, got %d", len(fc.msgs))
	}

	// The same Helper is reusable for another message.
	if err := h.Send(Message{To: []string{"two@example.com"}, Subject: "again", Body: "b2"}); err != nil {
		t.Fatalf("second Send: %v", err)
	}
	if len(fc.msgs) != 2 {
		t.Fatalf("expected 2 messages after reuse, got %d", len(fc.msgs))
	}
}

// TestSendRetriesThenFails proves Retry.Count drives re-dials: Count+1 attempts.
func TestSendRetriesThenFails(t *testing.T) {
	fc := &fakeClient{err: errors.New("boom")}
	h := &Helper{
		Client: fc,
		Options: Options{
			Sender: Sender{From: "a@b.co"},
			Retry:  Retry{Count: 2},
		},
	}

	if err := h.Send(Message{To: []string{"c@d.co"}}); err == nil {
		t.Fatal("expected an error after exhausting retries")
	}
	if len(fc.msgs) != 3 {
		t.Fatalf("expected 3 attempts (Count+1), got %d", len(fc.msgs))
	}
}

// TestNewHelperSendNoTLSNoAuth exercises the real go-mail client end-to-end
// against a server that advertises neither STARTTLS nor AUTH.
func TestNewHelperSendNoTLSNoAuth(t *testing.T) {
	s := newFakeSMTP(t)
	host, portStr, _ := net.SplitHostPort(s.addr)

	h, err := NewHelper(
		SenderHost(host),
		SenderPort(mustPort(t, portStr)),
		SenderFrom("noreply@bigstack.co"),
		SenderTLS(TLSNone),
		SenderAuth(false),
		RetryCount(0),
	)
	if err != nil {
		t.Fatalf("NewHelper: %v", err)
	}

	err = h.Send(Message{
		To:      []string{"user@example.com"},
		Subject: "[FET Cloud] User Account Activation",
		Body:    "Account / password: user@example.com/secret123",
	})
	if err != nil {
		t.Fatalf("Send over no-TLS/no-auth server failed: %v", err)
	}

	payload := <-s.got
	if !strings.Contains(payload, "Subject: [FET Cloud] User Account Activation") {
		t.Errorf("subject missing from message:\n%s", payload)
	}
	if !strings.Contains(payload, "secret123") {
		t.Errorf("body missing from message:\n%s", payload)
	}
}

func TestNewHelperInvalidTLS(t *testing.T) {
	if _, err := NewHelper(SenderHost("x"), SenderPort(25), SenderTLS("bogus")); err == nil {
		t.Fatal("expected NewHelper to reject an invalid TLS policy")
	}
}

// TestTLSValidate pins the exported check callers use to reject a bad policy at
// write time. Empty must stay valid — it is how a config written before the
// field existed asks for the default — so this also guards against a future
// tightening that would break stored configs.
func TestTLSValidate(t *testing.T) {
	for _, in := range []TLS{TLSNone, TLSOpportunistic, TLSMandatory, ""} {
		if err := in.Validate(); err != nil {
			t.Errorf("Validate(%q): unexpected error %v", in, err)
		}
	}

	for _, in := range []TLS{"starttls", "None", "MANDATORY", "true"} {
		if err := in.Validate(); err == nil {
			t.Errorf("Validate(%q): expected an error, got nil", in)
		}
	}
}

func TestTLSPolicyMapping(t *testing.T) {
	cases := map[TLS]bool{
		TLSNone:          true,
		TLSOpportunistic: true,
		TLSMandatory:     true,
		"":               true, // empty defaults to mandatory
		"bogus":          false,
	}
	for in, ok := range cases {
		_, err := in.policy()
		if ok && err != nil {
			t.Errorf("policy(%q): unexpected error %v", in, err)
		}
		if !ok && err == nil {
			t.Errorf("policy(%q): expected error, got nil", in)
		}
	}
}

// fakeSMTP is a minimal, single-connection SMTP server that advertises neither
// STARTTLS nor AUTH, so the no-TLS / no-auth path can be tested end-to-end.
type fakeSMTP struct {
	addr string
	ln   net.Listener
	got  chan string
}

func newFakeSMTP(t *testing.T) *fakeSMTP {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	s := &fakeSMTP{addr: ln.Addr().String(), ln: ln, got: make(chan string, 1)}
	go s.serve()
	t.Cleanup(func() { _ = ln.Close() })
	return s
}

func (s *fakeSMTP) serve() {
	conn, err := s.ln.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	r := bufio.NewReader(conn)
	w := func(line string) { _, _ = conn.Write([]byte(line + "\r\n")) }

	w("220 mail.test ESMTP ready")
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		cmd := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(cmd, "EHLO"):
			w("250-mail.test")
			w("250 8BITMIME")
		case strings.HasPrefix(cmd, "HELO"):
			w("250 mail.test")
		case strings.HasPrefix(cmd, "MAIL"), strings.HasPrefix(cmd, "RCPT"):
			w("250 2.1.0 OK")
		case strings.HasPrefix(cmd, "NOOP"), strings.HasPrefix(cmd, "RSET"):
			w("250 2.0.0 OK")
		case strings.HasPrefix(cmd, "DATA"):
			w("354 End data with <CR><LF>.<CR><LF>")
			var b strings.Builder
			for {
				dl, err := r.ReadString('\n')
				if err != nil {
					return
				}
				if strings.TrimRight(dl, "\r\n") == "." {
					break
				}
				b.WriteString(dl)
			}
			s.got <- b.String()
			w("250 2.0.0 OK: queued")
		case strings.HasPrefix(cmd, "QUIT"):
			w("221 2.0.0 Bye")
			return
		default:
			w("250 2.0.0 OK")
		}
	}
}

func mustPort(t *testing.T, s string) int {
	t.Helper()
	p := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			t.Fatalf("bad port %q", s)
		}
		p = p*10 + int(c-'0')
	}
	return p
}

// TestBuildMessageContentType proves Body is plaintext by default and switches
// to text/html only when Message.HTML is set, so existing plaintext callers
// (FET's notify mails) are unaffected by the HTML support.
func TestBuildMessageContentType(t *testing.T) {
	h := &Helper{Options: Options{Sender: Sender{From: "noreply@bigstack.co"}}}

	cases := []struct {
		name string
		html bool
		want string
	}{
		{"plaintext by default", false, "text/plain"},
		{"html when flagged", true, "text/html"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := h.buildMessage(Message{To: []string{"user@example.com"}, Body: "x", HTML: tc.html})
			if err != nil {
				t.Fatalf("buildMessage: %v", err)
			}
			var buf bytes.Buffer
			if _, err := m.WriteTo(&buf); err != nil {
				t.Fatalf("WriteTo: %v", err)
			}
			if !strings.Contains(buf.String(), "Content-Type: "+tc.want) {
				t.Errorf("expected Content-Type %s, got:\n%s", tc.want, buf.String())
			}
		})
	}
}

// TestSendHTMLWithInline exercises the real go-mail client end-to-end and
// proves an Inline part keeps the Content-ID its cid: URL resolves to — the
// shape the quota-alert and billing mails need (branded HTML + inline logo).
func TestSendHTMLWithInline(t *testing.T) {
	s := newFakeSMTP(t)
	host, portStr, _ := net.SplitHostPort(s.addr)

	h, err := NewHelper(
		SenderHost(host),
		SenderPort(mustPort(t, portStr)),
		SenderFrom("noreply@bigstack.co"),
		SenderTLS(TLSNone),
		SenderAuth(false),
		RetryCount(0),
	)
	if err != nil {
		t.Fatalf("NewHelper: %v", err)
	}

	err = h.Send(Message{
		To:      []string{"admin@example.com"},
		Subject: "Quota Alert: instances in demo",
		Body:    `<html><body><img src="cid:logo_cmp_light@cmp" /></body></html>`,
		HTML:    true,
		Inlines: []Inline{{
			ContentID:   "logo_cmp_light@cmp",
			ContentType: "image/png",
			Content:     []byte("\x89PNG\r\n\x1a\n"),
		}},
	})
	if err != nil {
		t.Fatalf("Send HTML with inline failed: %v", err)
	}

	payload := <-s.got
	for _, want := range []string{
		"Content-Type: text/html",
		"Content-Type: image/png",
		"logo_cmp_light@cmp",
	} {
		if !strings.Contains(payload, want) {
			t.Errorf("expected %q in message:\n%s", want, payload)
		}
	}
}

// TestBuildMessageInlineContentIDHeader pins the exact Content-ID header the
// body's cid: URL must match. jordan-wright/email (the origin package) derived
// it from the attachment file name; go-mail needs it set explicitly, so this
// guards the port from silently breaking every inline logo.
func TestBuildMessageInlineContentIDHeader(t *testing.T) {
	h := &Helper{Options: Options{Sender: Sender{From: "noreply@bigstack.co"}}}

	m, err := h.buildMessage(Message{
		To:      []string{"user@example.com"},
		Body:    `<img src="cid:logo_cmp_light@cmp" />`,
		HTML:    true,
		Inlines: []Inline{{ContentID: "logo_cmp_light@cmp", ContentType: "image/png", Content: []byte("png")}},
	})
	if err != nil {
		t.Fatalf("buildMessage: %v", err)
	}

	var buf bytes.Buffer
	if _, err := m.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	// Header names are case-insensitive; go-mail emits "Content-Id".
	if !strings.Contains(strings.ToLower(buf.String()), "content-id: <logo_cmp_light@cmp>") {
		t.Errorf("inline part is missing Content-ID: <logo_cmp_light@cmp>:\n%s", buf.String())
	}
}
