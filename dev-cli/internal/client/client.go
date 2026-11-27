// Package client implements the management API's HTTPS and response-envelope protocol.
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/canonical/microcloud-cluster-manager/dev-cli/internal/session"
)

// Error represents an HTTP or local failure with a stable CLI exit code.
type Error struct {
	Code    int
	Message string
}

// Error returns the user-facing diagnostic.
func (e *Error) Error() string { return e.Message }

// ExitCode classifies errors without terminating the process.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, context.Canceled) {
		return 130
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return 5
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return 1
}

// Response is the management API envelope, retaining metadata not yet known to this client.
type Response struct {
	Type       string          `json:"type"`
	Status     string          `json:"status"`
	StatusCode int             `json:"status_code"`
	Operation  string          `json:"operation"`
	Error      string          `json:"error"`
	ErrorCode  int             `json:"error_code"`
	Metadata   json.RawMessage `json:"metadata"`
}

// Options configures transport without loading backend configuration or private keys.
type Options struct {
	Server     string
	CACert     string
	Insecure   bool
	Timeout    time.Duration
	HTTPClient *http.Client
}

// Client sends requests only to one management API origin.
type Client struct {
	HTTP    *http.Client
	Origin  string
	Session *session.Store
	Timeout time.Duration
	// Persist is false during acquisition, until the supplied credentials are verified.
	Persist bool
}

// New creates a TLS client. An injected client is copied and still cannot follow redirects.
func New(o Options, s *session.Store) (*Client, error) {
	u, err := url.Parse(o.Server)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return nil, &Error{2, "--server must be an HTTPS origin without credentials, path, query, or fragment"}
	}
	if o.Timeout <= 0 {
		return nil, &Error{2, "--timeout must be positive"}
	}
	if o.CACert != "" && o.Insecure {
		return nil, &Error{2, "--insecure cannot be combined with a CA file from --ca-cert or MCM_DEV_CA_CERT"}
	}
	h := &http.Client{}
	if o.HTTPClient != nil {
		*h = *o.HTTPClient
	} else {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: o.Insecure} // Explicit development-only flag.
		if o.CACert != "" {
			pem, err := os.ReadFile(o.CACert)
			if err != nil {
				return nil, fmt.Errorf("read CA certificate: %w", err)
			}
			pool, err := x509.SystemCertPool()
			if err != nil || pool == nil {
				pool = x509.NewCertPool()
			}
			if !pool.AppendCertsFromPEM(pem) {
				return nil, fmt.Errorf("CA file contains no certificates")
			}
			t.TLSClientConfig.RootCAs = pool
		}
		h.Transport = t
	}
	h.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	return &Client{HTTP: h, Origin: strings.TrimSuffix(o.Server, "/"), Session: s, Timeout: o.Timeout, Persist: true}, nil
}

// Path constructs a versioned path, escaping each identifier independently.
func Path(segments ...string) string {
	p := "/1.0"
	for _, s := range segments {
		p += "/" + url.PathEscape(s)
	}
	return p
}

// Do sends one request, persists cookie changes, and decodes the API envelope. It never replays writes.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body any, authenticated bool) (*Response, error) {
	resp, err := c.request(ctx, method, path, query, body, authenticated)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	var envelope Response
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 16<<20))
	err = decoder.Decode(&envelope)
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return nil, fmt.Errorf("%s %s: %w", method, path, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || envelope.Type == "error" {
		code := 1
		if resp.StatusCode == 401 || envelope.ErrorCode == 401 {
			code = 3
		}
		if resp.StatusCode == 403 || envelope.ErrorCode == 403 {
			code = 4
		}
		message := fmt.Sprintf("%s %s: HTTP %d", method, path, resp.StatusCode)
		if envelope.Error != "" {
			message += ": " + envelope.Error
		}
		if code == 3 {
			message += "; log in in your browser and run auth login"
		}
		if method != http.MethodGet {
			message += "; inspect resource state before retrying the write"
		}
		return nil, &Error{code, message}
	}
	if err != nil || (envelope.Type != "sync" && envelope.Type != "async") {
		return nil, fmt.Errorf("%s %s: invalid management API response envelope", method, path)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("%s %s: %w", method, path, err)
		}
		return nil, fmt.Errorf("%s %s: unexpected trailing response data", method, path)
	}
	if len(envelope.Metadata) == 0 {
		envelope.Metadata = json.RawMessage("null")
	}
	return &envelope, nil
}

func (c *Client) request(ctx context.Context, method, path string, query url.Values, body any, authenticated bool) (*http.Response, error) {
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.ContainsAny(path, "?#") {
		return nil, fmt.Errorf("invalid API path")
	}
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	target := c.Origin + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	// The cancel function must live until the response body has been read.
	req, err := http.NewRequestWithContext(ctx, method, target, bytes.NewReader(data))
	if err != nil {
		cancel()
		return nil, err
	}
	req.Header.Set("User-Agent", "mcm-dev")
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if authenticated {
		if err := c.Session.Apply(req); err != nil {
			cancel()
			return nil, &Error{3, err.Error()}
		}
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("%s %s: %w", method, path, err)
	}
	resp.Body = &cancelBody{ReadCloser: resp.Body, cancel: cancel}
	if authenticated {
		changed, err := c.Session.Update(resp.Cookies())
		if err == nil && changed && c.Persist {
			err = c.Session.Save()
		}
		if err != nil {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("session persistence failed; server action may have succeeded: %w", err)
		}
	}
	return resp, nil
}

type cancelBody struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (b *cancelBody) Close() error { defer b.cancel(); return b.ReadCloser.Close() }

// Logout attempts server-side logout without following its redirect.
func (c *Client) Logout(ctx context.Context) error {
	resp, err := c.request(ctx, http.MethodGet, "/oidc/logout", nil, nil, true)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("remote logout: HTTP %d", resp.StatusCode)
	}
	return nil
}
