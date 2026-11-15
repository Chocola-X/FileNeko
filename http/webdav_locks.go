package fbhttp

import (
	"crypto/rand"
	"sync"
	"time"

	"golang.org/x/net/webdav"
)

const maxDAVLocks = 4096

type davLockEntry struct {
	internal string
	root     string
	expires  time.Time
	owner    uint
	virtual  string
}

// memLS issues sequential timestamp tokens. Never expose or accept those tokens
// at the HTTP boundary: map cryptographically random capabilities to them, and
// retain the root so refresh/unlock can authorize before mutating the lock.
type davLockRegistry struct {
	mu      sync.Mutex
	system  webdav.LockSystem
	entries map[string]davLockEntry
}

func newDAVLockRegistry() *davLockRegistry {
	return &davLockRegistry{system: webdav.NewMemLS(), entries: make(map[string]davLockEntry)}
}

func (s *davLockRegistry) collect(now time.Time) {
	for token, entry := range s.entries {
		if !entry.expires.IsZero() && !now.Before(entry.expires) {
			delete(s.entries, token)
		}
	}
}

func lockExpiry(now time.Time, duration time.Duration) time.Time {
	if duration < 0 {
		return time.Time{}
	}
	return now.Add(duration)
}

func (s *davLockRegistry) Create(now time.Time, details webdav.LockDetails, owner uint, virtual string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.collect(now)
	if len(s.entries) >= maxDAVLocks {
		return "", webdav.ErrLocked
	}
	internal, err := s.system.Create(now, details)
	if err != nil {
		return "", err
	}
	token := "urn:filebrowser:lock:" + rand.Text()
	s.entries[token] = davLockEntry{internal: internal, root: details.Root, expires: lockExpiry(now, details.Duration), owner: owner, virtual: virtual}
	return token, nil
}

func (s *davLockRegistry) Confirm(now time.Time, name0, name1 string, conditions ...webdav.Condition) (func(), error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.collect(now)
	translated := append([]webdav.Condition(nil), conditions...)
	for i, condition := range translated {
		// Unknown tokens must never pass through to the sequential inner store.
		translated[i].Token = s.entries[condition.Token].internal
	}
	return s.system.Confirm(now, name0, name1, translated...)
}

func (s *davLockRegistry) Refresh(now time.Time, token string, duration time.Duration, resolve func(davLockEntry) (string, error)) (webdav.LockDetails, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.collect(now)
	entry, ok := s.entries[token]
	if !ok {
		return webdav.LockDetails{}, webdav.ErrNoSuchLock
	}
	virtual, err := resolve(entry)
	if err != nil {
		return webdav.LockDetails{}, webdav.ErrForbidden
	}
	details, err := s.system.Refresh(now, entry.internal, duration)
	if err == nil {
		entry.expires = lockExpiry(now, duration)
		s.entries[token] = entry
		details.Root = virtual
	}
	return details, err
}

func (s *davLockRegistry) Unlock(now time.Time, token string, allowed func(davLockEntry) bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.collect(now)
	entry, ok := s.entries[token]
	if !ok {
		return webdav.ErrNoSuchLock
	}
	if !allowed(entry) {
		return webdav.ErrForbidden
	}
	if err := s.system.Unlock(now, entry.internal); err != nil {
		return err
	}
	delete(s.entries, token)
	return nil
}
