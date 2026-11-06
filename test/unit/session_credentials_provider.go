package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/canonical/microcloud-cluster-manager/internal/app/management-api/core/auth"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
	"github.com/google/uuid"
)

// newSessionCredentialsProvider returns a provider holding the access token. It has no verifier, so any attempt
// to refresh the tokens panics.
func newSessionCredentialsProvider(accessToken string) (*auth.SessionCredentialsProvider, error) {
	secret := strings.Repeat("a", 32)
	tokens, err := auth.EncryptTokens("id-token", accessToken, "refresh-token", secret)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt tokens: %w", err)
	}

	return auth.NewSessionCredentialsProvider(nil, uuid.Nil, secret, tokens.AccessToken), nil
}

// refreshSessionTokens refreshes the tokens of the provider, rejecting the given token. It returns an error instead
// of panicking if the provider tried to redeem the refresh token, which it cannot without a verifier.
func refreshSessionTokens(provider *auth.SessionCredentialsProvider, rejectedToken string) (token string, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("tried to refresh the tokens: %v", r)
		}
	}()

	return provider.RefreshSessionTokens(context.Background(), rejectedToken)
}

func testSessionCredentialsProvider_RefreshReturnsCurrentToken() (testName string, testFunc func(t *testing.T)) {
	return "session credentials provider does not refresh a token that was not rejected", func(t *testing.T) {
		var condition string

		provider, err := newSessionCredentialsProvider("current-token")
		helpers.LogTestOutcome(t, "Should set up a provider", err)

		{
			condition = "Should return the current token when a different token was rejected"

			var token string
			token, err = refreshSessionTokens(provider, "rejected-token")
			if err == nil && token != "current-token" {
				err = fmt.Errorf("got %q", token)
			}

			helpers.LogTestOutcome(t, condition, err)
		}

		{
			condition = "Should keep the current token"

			var token string
			token, err = provider.GetAccessToken()
			if err == nil && token != "current-token" {
				err = fmt.Errorf("got %q", token)
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

func testSessionCredentialsProvider_ConcurrentRefreshReturnsCurrentToken() (testName string, testFunc func(t *testing.T)) {
	return "session credentials provider shares the current token between concurrent refreshes", func(t *testing.T) {
		var condition string

		provider, err := newSessionCredentialsProvider("new-token")
		helpers.LogTestOutcome(t, "Should set up a provider", err)

		{
			condition = "Should return the current token to every request that was rejected with the old token"

			const requests = 16
			errs := make([]error, requests)

			var wg sync.WaitGroup
			for i := range requests {
				wg.Go(func() {
					token, err := refreshSessionTokens(provider, "old-token")
					if err == nil && token != "new-token" {
						err = fmt.Errorf("got %q", token)
					}

					errs[i] = err
				})
			}

			wg.Wait()

			helpers.LogTestOutcome(t, condition, firstError(errs))
		}
	}
}

// firstError returns the first non-nil error.
func firstError(errs []error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}
