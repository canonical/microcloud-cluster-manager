// Package session stores the browser cookies used by the development CLI.
package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// Names are the only cookies the CLI imports or persists.
var Names = []string{"oidc_session", "user_secret"}

// Store is an origin-bound, versioned browser session. It must not be shared by concurrent CLI processes.
type Store struct {
	Version int            `json:"version"`
	Origin  string         `json:"origin"`
	Cookies []*http.Cookie `json:"cookies"`
	file    string
}

// DefaultPath returns the user configuration location for CLI credentials.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mcm-dev", "session.json"), nil
}

// Load reads a session, rejecting unexpected permissions, formats, or server origins.
func Load(file, origin string) (*Store, error) {
	info, err := os.Lstat(file)
	if errors.Is(err, os.ErrNotExist) {
		return &Store{Version: 1, Origin: origin, file: file}, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return nil, fmt.Errorf("session file must be a regular file with permissions 0600")
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	s := &Store{file: file}
	if err := json.Unmarshal(b, s); err != nil {
		return nil, fmt.Errorf("invalid session file")
	}
	if s.Version != 1 {
		return nil, fmt.Errorf("unsupported session file version")
	}
	if s.Origin == "" || s.Origin != origin {
		return nil, fmt.Errorf("session belongs to a different server; use a separate --session-file")
	}
	for _, c := range s.Cookies {
		if c == nil || !known(c.Name) || c.Valid() != nil {
			return nil, fmt.Errorf("invalid session cookie")
		}
	}
	return s, nil
}

func known(name string) bool {
	return name == "oidc_session" || name == "user_secret"
}

// Set imports cookie values without inventing a browser expiry; the server validates their lifetime.
func (s *Store) Set(values []string) error {
	if len(values) != 2 {
		return fmt.Errorf("expected two cookie values")
	}
	s.Cookies = nil
	for i, value := range values {
		c := &http.Cookie{Name: Names[i], Value: value, Path: "/", Secure: true, HttpOnly: true}
		if value == "" || c.Valid() != nil {
			return fmt.Errorf("invalid %s cookie value", Names[i])
		}
		s.Cookies = append(s.Cookies, c)
	}
	return nil
}

// Apply adds both valid cookies to a request. JWT expiry is intentionally left to the server.
func (s *Store) Apply(r *http.Request) error {
	if r.URL.Scheme+"://"+r.URL.Host != s.Origin {
		return fmt.Errorf("refusing to send session to a different origin")
	}
	for _, name := range Names {
		var found *http.Cookie
		for _, c := range s.Cookies {
			if c.Name == name {
				found = c
			}
		}
		if found == nil || found.MaxAge < 0 || (!found.Expires.IsZero() && !found.Expires.After(time.Now())) {
			return fmt.Errorf("missing or expired %s; log in in your browser and run auth login", name)
		}
		r.AddCookie(found)
	}
	return nil
}

// Update captures cookie renewal and deletion, including on error responses.
func (s *Store) Update(cookies []*http.Cookie) (bool, error) {
	changed := false
	origin, _ := url.Parse(s.Origin)
	for _, c := range cookies {
		if !known(c.Name) {
			continue
		}
		if c.Domain != "" && c.Domain != origin.Hostname() {
			return changed, fmt.Errorf("unexpected session cookie domain")
		}
		if c.Path != "" && c.Path != "/" {
			return changed, fmt.Errorf("unexpected session cookie path")
		}
		changed = true
		remaining := make([]*http.Cookie, 0, 2)
		for _, old := range s.Cookies {
			if old.Name != c.Name {
				remaining = append(remaining, old)
			}
		}
		if c.MaxAge >= 0 && (c.Expires.IsZero() || c.Expires.After(time.Now())) {
			if c.MaxAge > 0 {
				c.Expires = time.Now().Add(time.Duration(c.MaxAge) * time.Second)
			}
			remaining = append(remaining, c)
		}
		s.Cookies = remaining
	}
	return changed, nil
}

// Save atomically replaces the private credential file.
func (s *Store) Save() error {
	dir := filepath.Dir(s.file)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode().Perm() != 0700 {
		return fmt.Errorf("session directory %q must be a directory with permissions 0700", dir)
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".session-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if _, err := f.Write(append(b, '\n')); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), s.file)
}

// Remove deletes local credentials, including when remote logout fails.
func (s *Store) Remove() error {
	s.Cookies = nil
	err := os.Remove(s.file)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
