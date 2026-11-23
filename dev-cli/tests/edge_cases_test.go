package tests

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/canonical/microcloud-cluster-manager/dev-cli/internal/client"
	"github.com/canonical/microcloud-cluster-manager/dev-cli/internal/command"
	"github.com/canonical/microcloud-cluster-manager/dev-cli/internal/session"
)

func TestCACertEnvironmentDefault(t *testing.T) {
	dir := t.TempDir()
	envCA := filepath.Join(dir, "env-ca.crt")
	flagCA := filepath.Join(dir, "flag-ca.crt")
	t.Setenv("MCM_DEV_CA_CERT", envCA)
	for _, tc := range []struct {
		name string
		args []string
		path string
		code int
	}{
		{name: "environment default", path: envCA, code: 1},
		{name: "explicit override", args: []string{"--ca-cert", flagCA}, path: flagCA, code: 1},
		{name: "empty override", args: []string{"--ca-cert="}, code: 3},
		{name: "conflict with insecure", args: []string{"--insecure"}, code: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := command.New(command.Options{Out: io.Discard, Err: io.Discard})
			c.SetArgs(append([]string{"--session-file", filepath.Join(dir, "session.json"), "query", "/1.0"}, tc.args...))
			err := c.Execute()
			if code := client.ExitCode(err); code != tc.code {
				t.Fatalf("exit code = %d, want %d: %v", code, tc.code, err)
			}
			var pathErr *os.PathError
			if tc.path != "" && (!errors.As(err, &pathErr) || pathErr.Path != tc.path) {
				t.Fatalf("expected error opening %q, got %v", tc.path, err)
			}
		})
	}
}

func TestMissingOrCorruptSession(t *testing.T) {
	f := setup(t, func(w http.ResponseWriter, r *http.Request) { t.Error("unusable credentials sent to server") })
	if err := os.Remove(f.file); err != nil {
		t.Fatal(err)
	}
	if _, err := f.run(t, "query", "/1.0"); client.ExitCode(err) != 3 {
		t.Fatalf("missing session: %v", err)
	}
	if _, err := f.run(t, "auth", "logout"); client.ExitCode(err) != 3 {
		t.Fatalf("logout missing session: %v", err)
	}
	if err := os.WriteFile(f.file, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := f.run(t, "query", "/1.0"); err == nil {
		t.Fatal("corrupt session accepted")
	}
}

func TestSessionOriginExpiryAndPermissions(t *testing.T) {
	file := filepath.Join(t.TempDir(), "mcm-dev", "session.json")
	s, err := session.Load(file, "https://one.test")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set([]string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	if _, err := session.Load(file, "https://two.test"); err == nil {
		t.Fatal("origin mismatch accepted")
	}
	req, _ := http.NewRequest("GET", "https://two.test/1.0", nil)
	if err := s.Apply(req); err == nil || req.Header.Get("Cookie") != "" {
		t.Fatal("credentials escaped origin")
	}
	s.Cookies[0].Expires = time.Now().Add(-time.Second)
	req, _ = http.NewRequest("GET", "https://one.test/1.0", nil)
	if err := s.Apply(req); err == nil {
		t.Fatal("expired cookie sent")
	}
	if err := os.Chmod(file, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := session.Load(file, "https://one.test"); err == nil {
		t.Fatal("public credential file accepted")
	}
}

func TestSessionRequiresVersionAndOrigin(t *testing.T) {
	const origin = "https://one.test"
	for _, test := range []struct {
		name string
		data string
	}{
		{name: "missing version", data: `{"origin":"https://one.test"}`},
		{name: "missing origin", data: `{"version":1}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "session.json")
			if err := os.WriteFile(file, []byte(test.data), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := session.Load(file, origin); err == nil {
				t.Fatalf("session accepted with %s", test.name)
			}
		})
	}
}

func TestSessionSaveRejectsInsecureDirectory(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "mcm-dev")
	file := filepath.Join(dir, "session.json")
	s, err := session.Load(file, "https://one.test")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []os.FileMode{0755, 0777} {
		if err := os.Chmod(dir, mode); err != nil {
			t.Fatal(err)
		}
		if err := s.Save(); err == nil {
			t.Fatalf("saved session in directory with mode %04o", mode)
		}
	}
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(parent, "linked")
	if err := os.Symlink(dir, link); err != nil {
		t.Fatal(err)
	}
	s, err = session.Load(filepath.Join(link, "session.json"), "https://one.test")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Save(); err == nil {
		t.Fatal("saved session through a symlinked directory")
	}
}

func TestInvalidLoginDoesNotReplaceExistingSession(t *testing.T) {
	f := setup(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = io.WriteString(w, `{"type":"error","error":"unauthorized"}`)
	})
	before, err := os.ReadFile(f.file)
	if err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	c := command.New(command.Options{In: strings.NewReader("new-session\nnew-secret\n"), Out: &out, Err: &stderr, HTTPClient: f.httpClient})
	c.SetArgs([]string{"--server", f.server.URL, "--session-file", f.file, "auth", "login", "--stdin"})
	if err := c.Execute(); client.ExitCode(err) != 3 {
		t.Fatalf("invalid login: %v", err)
	}
	after, err := os.ReadFile(f.file)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("failed login replaced old credentials")
	}
	if strings.Contains(out.String()+stderr.String(), "new-secret") {
		t.Fatal("cookie leaked")
	}
}

func TestHiddenPromptAndStdinValidation(t *testing.T) {
	f := setup(t, func(w http.ResponseWriter, r *http.Request) { sync(w, map[string]string{"email": "dev@example.org"}) })
	for _, input := range []string{"one\n", "one\ntwo\nthree\n"} {
		var out bytes.Buffer
		c := command.New(command.Options{In: strings.NewReader(input), Out: &out, Err: io.Discard, HTTPClient: f.httpClient})
		c.SetArgs([]string{"--server", f.server.URL, "--session-file", f.file, "auth", "login", "--stdin"})
		if err := c.Execute(); client.ExitCode(err) != 2 {
			t.Fatalf("invalid stdin accepted: %v", err)
		}
	}
	var names []string
	c := command.New(command.Options{In: strings.NewReader(""), Out: io.Discard, Err: io.Discard, HTTPClient: f.httpClient, ReadSecret: func(name string) (string, error) { names = append(names, name); return "hidden-value", nil }})
	c.SetArgs([]string{"--server", f.server.URL, "--session-file", f.file, "auth", "login"})
	if err := c.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.Join(names, ",") != "oidc_session,user_secret" {
		t.Fatalf("wrong prompts: %v", names)
	}
}

func TestMalformedEnvelopesAndRedirect(t *testing.T) {
	for _, body := range []string{`<html>login</html>`, `{}`, `{"type":"unknown"}`, `{"type":"sync","metadata":null} {}`} {
		t.Run(body, func(t *testing.T) {
			f := setup(t, func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, body) })
			_, err := f.run(t, "query", "/1.0/remote-cluster")
			if client.ExitCode(err) != 1 {
				t.Fatalf("bad envelope: %v", err)
			}
		})
	}
	calls := 0
	f := setup(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Location", "https://foreign.test/login")
		w.WriteHeader(302)
	})
	if _, err := f.run(t, "query", "/1.0/remote-cluster"); client.ExitCode(err) != 1 {
		t.Fatalf("redirect: %v", err)
	}
	if calls != 1 {
		t.Fatal("redirect followed")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestRequestDeadlineAndCancellation(t *testing.T) {
	f := setup(t, func(w http.ResponseWriter, r *http.Request) { sync(w, nil) })
	h := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() })}
	for _, cancelled := range []bool{false, true} {
		c := command.New(command.Options{In: strings.NewReader(""), Out: io.Discard, Err: io.Discard, HTTPClient: h})
		c.SetArgs([]string{"--server", f.server.URL, "--session-file", f.file, "--timeout=5ms", "query", "/1.0/remote-cluster"})
		ctx, cancel := context.WithCancel(context.Background())
		if cancelled {
			cancel()
		}
		err := c.ExecuteContext(ctx)
		cancel()
		want := 5
		if cancelled {
			want = 130
		}
		if client.ExitCode(err) != want {
			t.Fatalf("error %v: want %d got %d", err, want, client.ExitCode(err))
		}
	}
}

func TestOperationIdentityValidation(t *testing.T) {
	for _, r := range []*client.Response{
		{Operation: "/unexpected/op", Metadata: json.RawMessage(`{}`)},
		{Operation: "/1.0/operations/a", Metadata: json.RawMessage(`{"id":"b"}`)},
		{Operation: "/1.0/operations/a%2Fb", Metadata: json.RawMessage(`{}`)},
	} {
		if _, err := client.OperationID(r); err == nil {
			t.Fatalf("unsafe operation accepted: %+v", r)
		}
	}
}

func TestHelpAndUnknownCommand(t *testing.T) {
	root := command.New(command.Options{Out: io.Discard, Err: io.Discard})
	commands := root.Commands()
	if len(commands) != 2 || commands[0].Name() != "auth" || commands[1].Name() != "query" {
		t.Fatal("only auth and query should be registered")
	}
	for _, args := range [][]string{{"--help"}, {"query", "--help"}, {"auth"}} {
		var out bytes.Buffer
		c := command.New(command.Options{Out: &out, Err: io.Discard})
		c.SetArgs(append([]string{"--session-file", "/does/not/exist/session.json"}, args...))
		if err := c.Execute(); err != nil || !strings.Contains(out.String(), "Usage:") {
			t.Fatalf("help failed: %v %s", err, out.String())
		}
	}
	for _, args := range [][]string{{"unknown"}, {"auth", "unknown"}, {"--unknown"}, {"status"}, {"configuration"}, {"cluster"}, {"enrollment-token"}, {"project"}, {"cluster-link"}, {"replicator"}, {"operation"}, {"--output=json", "query", "/1.0"}} {
		c := command.New(command.Options{Out: io.Discard, Err: io.Discard})
		c.SetArgs(args)
		if err := c.Execute(); client.ExitCode(err) != 2 {
			t.Fatalf("usage error classification: %v", err)
		}
	}
	if client.ExitCode(errors.New("failure")) != 1 {
		t.Fatal("unclassified error must fail")
	}
}

func TestTLSConfigurationAndOriginValidation(t *testing.T) {
	for _, server := range []string{"http://one.test", "https://one.test/1.0", "https://user:pass@one.test", "https://one.test?query=x", "https://one.test#fragment"} {
		if _, err := client.New(client.Options{Server: server, Timeout: time.Second}, nil); client.ExitCode(err) != 2 {
			t.Fatalf("accepted invalid server %q: %v", server, err)
		}
	}
	for _, insecure := range []bool{false, true} {
		c, err := client.New(client.Options{Server: "https://one.test", Timeout: time.Second, Insecure: insecure}, nil)
		if err != nil {
			t.Fatal(err)
		}
		transport := c.HTTP.Transport.(*http.Transport)
		if transport.TLSClientConfig.InsecureSkipVerify != insecure {
			t.Fatal("TLS verification configuration ignored")
		}
		transport.CloseIdleConnections()
	}
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, template, template, public, private)
	if err != nil {
		t.Fatal(err)
	}
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	file := filepath.Join(t.TempDir(), "ca.crt")
	if err := os.WriteFile(file, cert, 0600); err != nil {
		t.Fatal(err)
	}
	c, err := client.New(client.Options{Server: "https://one.test", Timeout: time.Second, CACert: file}, nil)
	if err != nil {
		t.Fatal(err)
	}
	transport := c.HTTP.Transport.(*http.Transport)
	if transport.TLSClientConfig.RootCAs == nil {
		t.Fatal("CA ignored")
	}
	transport.CloseIdleConnections()
	if err := os.WriteFile(file, []byte("not a certificate"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := client.New(client.Options{Server: "https://one.test", Timeout: time.Second, CACert: file}, nil); err == nil {
		t.Fatal("invalid CA accepted")
	}
}
