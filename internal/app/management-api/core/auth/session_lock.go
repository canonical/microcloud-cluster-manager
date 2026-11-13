package auth

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

// expiredSessionRenewalGracePeriod is how long the result of renewing an expired session is kept, so that requests
// carrying the same stale session cookie reuse the renewed session instead of failing once the old one is deleted.
const expiredSessionRenewalGracePeriod = time.Minute

// sessionLocks provides per-session mutual exclusion. Waiting for a lock respects context cancellation.
type sessionLocks struct {
	mu    sync.Mutex
	locks map[uuid.UUID]*sessionLock
}

type sessionLock struct {
	ch   chan struct{}
	refs int
}

// lock acquires the lock for the given session ID. The returned function releases it.
func (l *sessionLocks) lock(ctx context.Context, sessionID uuid.UUID) (func(), error) {
	l.mu.Lock()
	if l.locks == nil {
		l.locks = make(map[uuid.UUID]*sessionLock)
	}

	entry, ok := l.locks[sessionID]
	if !ok {
		entry = &sessionLock{ch: make(chan struct{}, 1)}
		l.locks[sessionID] = entry
	}

	entry.refs++
	l.mu.Unlock()

	release := func() {
		l.mu.Lock()
		entry.refs--
		if entry.refs == 0 {
			delete(l.locks, sessionID)
		}

		l.mu.Unlock()
	}

	select {
	case entry.ch <- struct{}{}:
	case <-ctx.Done():
		release()
		return nil, ctx.Err()
	}

	return func() {
		<-entry.ch
		release()
	}, nil
}

// renewedSession is the outcome of renewing an expired session.
type renewedSession struct {
	claims               *oidc.IDTokenClaims
	sessionID            uuid.UUID
	userSecret           string
	encryptedAccessToken string
	expiry               time.Time
	validUntil           time.Time
}

// renewedSessions caches renewed sessions keyed by the expired session ID they replaced.
type renewedSessions struct {
	mu      sync.Mutex
	entries map[uuid.UUID]renewedSession
}

func (c *renewedSessions) get(oldSessionID uuid.UUID) (renewedSession, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[oldSessionID]
	if !ok || time.Now().After(entry.validUntil) {
		return renewedSession{}, false
	}

	return entry, true
}

func (c *renewedSessions) set(oldSessionID uuid.UUID, entry renewedSession) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.entries == nil {
		c.entries = make(map[uuid.UUID]renewedSession)
	}

	now := time.Now()
	for id, e := range c.entries {
		if now.After(e.validUntil) {
			delete(c.entries, id)
		}
	}

	entry.validUntil = now.Add(expiredSessionRenewalGracePeriod)
	c.entries[oldSessionID] = entry
}
