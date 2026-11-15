package storage

import (
	"errors"
	"time"
)

var ErrInvalidSession = errors.New("invalid or consumed session")

// SessionStore persists revocable random session identifiers. JWT payloads are
// only client metadata; a valid signature alone no longer authorizes a request.
type SessionStore interface {
	Create(id string, userID uint, expires time.Time) error
	Valid(id string, userID uint, now time.Time) (bool, error)
	Delete(id string) error
	Replace(oldID, newID string, userID uint, expires time.Time) error
}
