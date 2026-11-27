package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/canonical/microcloud-cluster-manager/dev-cli/internal/client"
	"github.com/canonical/microcloud-cluster-manager/dev-cli/internal/command"
	"github.com/canonical/microcloud-cluster-manager/dev-cli/internal/session"
)

type fixture struct {
	server     *httptest.Server
	httpClient *http.Client
	file       string
}

type handlerTransport struct{ handler http.Handler }

func (h handlerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if err := r.Context().Err(); err != nil {
		return nil, err
	}
	w := httptest.NewRecorder()
	h.handler.ServeHTTP(w, r)
	response := w.Result()
	response.Request = r
	return response, nil
}

func setup(t *testing.T, handler http.HandlerFunc) *fixture {
	t.Helper()
	// Most contract tests need no sockets. Opt in to real TLS transport when the
	// environment allows loopback listeners: MCM_DEV_TLS_TESTS=1 go test ./dev-cli/...
	srv := &httptest.Server{URL: "https://management.test"}
	httpClient := &http.Client{Transport: handlerTransport{handler}}
	if os.Getenv("MCM_DEV_TLS_TESTS") == "1" {
		srv = httptest.NewTLSServer(handler)
		t.Cleanup(srv.Close)
		httpClient = srv.Client()
	}
	file := filepath.Join(t.TempDir(), "credentials", "session.json")
	s, err := session.Load(file, srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Set([]string{"session-value", "secret-value"}); err != nil {
		t.Fatal(err)
	}
	if err = s.Save(); err != nil {
		t.Fatal(err)
	}
	return &fixture{server: srv, httpClient: httpClient, file: file}
}

func (f *fixture) run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out, stderr bytes.Buffer
	c := command.New(command.Options{Out: &out, Err: &stderr, In: strings.NewReader(""), HTTPClient: f.httpClient})
	c.SetArgs(append([]string{"--server", f.server.URL, "--session-file", f.file}, args...))
	err := c.ExecuteContext(context.Background())
	return out.String(), err
}

func sync(w http.ResponseWriter, metadata any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"type": "sync", "status": "Success", "status_code": 200, "metadata": metadata})
}

func TestLoginRenewalAndLogout(t *testing.T) {
	calls := 0
	f := setup(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path == "/oidc/logout" {
			http.Redirect(w, r, "https://should-never-be-followed.invalid/", http.StatusFound)
			return
		}
		if r.URL.Path != "/1.0" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		cookie, _ := r.Cookie("oidc_session")
		if cookie == nil || cookie.Value != "pasted-session" {
			t.Errorf("bad imported cookie")
		}
		for _, name := range session.Names {
			http.SetCookie(w, &http.Cookie{Name: name, Value: "renewed-" + name, Path: "/", Secure: true, Expires: time.Now().Add(time.Hour)})
		}
		sync(w, map[string]any{"email": "dev@example.org", "groups": []string{"admins"}, "trusted": true})
	})
	var out, stderr bytes.Buffer
	c := command.New(command.Options{In: strings.NewReader("pasted-session\npasted-secret\n"), Out: &out, Err: &stderr, HTTPClient: f.httpClient})
	c.SetArgs([]string{"--server", f.server.URL, "--session-file", f.file, "auth", "login", "--stdin"})
	if err := c.Execute(); err != nil {
		t.Fatal(err)
	}
	s, err := session.Load(f.file, f.server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Cookies) != 2 || s.Cookies[0].Value != "renewed-oidc_session" {
		t.Fatalf("did not save rotation: %+v", s.Cookies)
	}
	for _, p := range []string{f.file, filepath.Dir(f.file)} {
		info, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm()&0077 != 0 {
			t.Errorf("insecure permissions %s", info.Mode())
		}
	}
	if _, err := f.run(t, "auth", "logout"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(f.file); !os.IsNotExist(err) {
		t.Fatalf("session file remains: %v", err)
	}
	if calls != 2 {
		t.Fatalf("redirect followed, calls=%d", calls)
	}
}

func TestErrorsAndNoWriteReplay(t *testing.T) {
	for _, status := range []int{401, 403, 404, 409, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			calls := 0
			f := setup(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				http.SetCookie(w, &http.Cookie{Name: "oidc_session", Value: "", Path: "/", Expires: time.Unix(1, 0)})
				w.WriteHeader(status)
				_ = json.NewEncoder(w).Encode(map[string]any{"type": "error", "error": "expected failure", "error_code": status})
			})
			_, err := f.run(t, "query", "/1.0/remote-cluster/lab", "-X", "DELETE")
			want := 1
			if status == 401 {
				want = 3
			}
			if status == 403 {
				want = 4
			}
			if client.ExitCode(err) != want || !strings.Contains(err.Error(), "expected failure") {
				t.Fatalf("error=%v code=%d", err, client.ExitCode(err))
			}
			if calls != 1 {
				t.Fatalf("write replayed %d times", calls)
			}
			s, e := session.Load(f.file, f.server.URL)
			if e != nil {
				t.Fatal(e)
			}
			if len(s.Cookies) != 1 {
				t.Fatal("cookie deletion on failed response not persisted")
			}
		})
	}
}
