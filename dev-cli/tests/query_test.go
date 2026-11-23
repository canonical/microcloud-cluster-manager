package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/canonical/microcloud-cluster-manager/dev-cli/internal/client"
	"github.com/canonical/microcloud-cluster-manager/dev-cli/internal/command"
)

func TestQueryRequestMapping(t *testing.T) {
	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
		t.Run(method, func(t *testing.T) {
			calls := 0
			f := setup(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != method || r.URL.EscapedPath() != "/1.0/remote-cluster/a%2Fb/replicators" {
					t.Errorf("request: %s %s", r.Method, r.URL.EscapedPath())
				}
				if values := r.URL.Query()["project"]; len(values) != 2 || values[0] != "p & q" || values[1] != "other" {
					t.Errorf("query lost: %v", r.URL.Query())
				}
				for _, name := range []string{"oidc_session", "user_secret"} {
					if _, err := r.Cookie(name); err != nil {
						t.Error(err)
					}
				}
				body, _ := io.ReadAll(r.Body)
				if string(body) != `{"large":18446744073709551615,"description":"a,b=c"}` {
					t.Errorf("body changed: %s", body)
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Error("JSON content type missing")
				}
				sync(w, map[string]any{"future": true})
			})
			out, err := f.run(t, "query", "/1.0/remote-cluster/a%2Fb/replicators?project=p+%26+q&project=other", "-X", method, "-d", `{"large":18446744073709551615,"description":"a,b=c"}`)
			if err != nil {
				t.Fatal(err)
			}
			if calls != 1 || !strings.Contains(out, `"future": true`) {
				t.Fatalf("calls %d output %s", calls, out)
			}
		})
	}
}

func TestQueryDefaultGetAndJSON(t *testing.T) {
	f := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/1.0/remote-cluster" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if len(body) > 0 {
			t.Errorf("unexpected body %s", body)
		}
		_, _ = io.WriteString(w, `{"type":"sync","status_code":200,"metadata":[{"name":"z","large":18446744073709551615},{"name":"a"}]}`)
	})
	out, err := f.run(t, "query", "/1.0/remote-cluster")
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid([]byte(out)) || !strings.Contains(out, "18446744073709551615") || strings.Index(out, `"z"`) > strings.Index(out, `"a"`) {
		t.Fatalf("response changed: %s", out)
	}
}

func TestQueryRawAndAsync(t *testing.T) {
	f := setup(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = io.WriteString(w, `{"type":"async","status":"Operation created","status_code":100,"operation":"/1.0/operations/op","metadata":{"id":"op","future":true}}`)
	})
	out, err := f.run(t, "query", "/1.0/remote-cluster/a/projects/p/state", "--raw")
	if err != nil {
		t.Fatal(err)
	}
	var envelope client.Response
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Type != "async" || envelope.Operation != "/1.0/operations/op" || !strings.Contains(string(envelope.Metadata), "future") {
		t.Fatalf("envelope lost: %s", out)
	}
	out, err = f.run(t, "query", "/1.0/remote-cluster/a/projects/p/state")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, `"metadata"`) || !strings.Contains(out, `"id": "op"`) {
		t.Fatalf("metadata output: %s", out)
	}
}

func TestQueryStdin(t *testing.T) {
	f := setup(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.Method != "PATCH" || string(body) != `{"description":"From stdin"}` {
			t.Errorf("request: %s %s", r.Method, body)
		}
		sync(w, nil)
	})
	var out bytes.Buffer
	c := command.New(command.Options{In: strings.NewReader("{\n\"description\":\"From stdin\"\n}"), Out: &out, Err: io.Discard, HTTPClient: f.httpClient})
	c.SetArgs([]string{"--server", f.server.URL, "--session-file", f.file, "query", "/1.0/remote-cluster/a", "-X", "PATCH", "-d", "-"})
	if err := c.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.String()) != "null" {
		t.Fatalf("empty result: %s", out.String())
	}
}

func TestQueryValidation(t *testing.T) {
	f := setup(t, func(w http.ResponseWriter, r *http.Request) { t.Error("invalid query reached server") })
	for _, args := range [][]string{
		{"query"}, {"query", "/1.0", "extra"}, {"query", "1.0/remote-cluster"},
		{"query", "https://foreign.test/1.0"}, {"query", "//foreign.test/1.0"}, {"query", "/1.0#fragment"},
		{"query", "/1.0?project=%xx"}, {"query", "/1.0", "-X", "TRACE"},
		{"query", "/1.0", "-d", "{"}, {"query", "/1.0", "-d", ""}, {"query", "/1.0", "-d", "{} {}"},
		{"query", "/1.0", "--wait"}, {"query", "/1.0/remote-cluster/", "--wait"},
		{"query", "/1.0/remote-cluster/a", "--wait-timeout=0s"},
		{"query", "/1.0", "--timeout=0s"},
		{"query", "/1.0", "--ca-cert=unused", "--insecure"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			_, err := f.run(t, args...)
			if client.ExitCode(err) != 2 {
				t.Fatalf("expected usage error, got %v", err)
			}
		})
	}
}

func TestQueryWait(t *testing.T) {
	for _, failed := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "failure"}[failed], func(t *testing.T) {
			calls := 0
			f := setup(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method == "PUT" {
					_, _ = io.WriteString(w, `{"type":"async","status_code":100,"operation":"https://foreign.test/1.0/operations/op","metadata":{"id":"op"}}`)
					return
				}
				if r.URL.EscapedPath() != "/1.0/remote-cluster/a%2Fb/operations/op" {
					t.Errorf("wrong polling path: %s", r.URL.EscapedPath())
				}
				code := 200
				if failed {
					code = 400
				}
				sync(w, map[string]any{"id": "op", "status_code": code, "future": true})
			})
			out, err := f.run(t, "query", "/1.0/remote-cluster/a%2Fb/projects/p/state", "-X", "PUT", "-d", `{"replica_mode":"leader"}`, "--wait", "--raw")
			if (!failed && err != nil) || (failed && client.ExitCode(err) != 1) {
				t.Fatalf("wait error: %v", err)
			}
			if calls != 2 || !strings.Contains(out, `"future": true`) || !strings.Contains(out, `"type": "sync"`) {
				t.Fatalf("final response: calls %d output %s", calls, out)
			}
		})
	}
}

func TestQueryFailedWriteIsNotRetried(t *testing.T) {
	calls := 0
	f := setup(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(403)
		_, _ = io.WriteString(w, `{"type":"error","error":"denied"}`)
	})
	_, err := f.run(t, "query", "/1.0/remote-cluster/a", "-X", "DELETE")
	if calls != 1 || client.ExitCode(err) != 4 {
		t.Fatalf("calls %d error %v", calls, err)
	}
}

func TestQueryWaitTimeout(t *testing.T) {
	f := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			_, _ = io.WriteString(w, `{"type":"async","operation":"/1.0/operations/op","metadata":{"id":"op"}}`)
			return
		}
		sync(w, map[string]any{"id": "op", "status_code": 103})
	})
	_, err := f.run(t, "query", "/1.0/remote-cluster/a/projects/p/state", "-X", "PUT", "--wait", "--wait-timeout=20ms")
	if client.ExitCode(err) != 5 || !strings.Contains(err.Error(), "op") {
		t.Fatalf("expected wait timeout with operation ID, got %v", err)
	}
}
