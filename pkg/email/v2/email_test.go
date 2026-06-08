package email

import (
	"bufio"
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
