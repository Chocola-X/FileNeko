package bolt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/storage"
	bolt "go.etcd.io/bbolt"
)

var sessionBucket = []byte("sessions")

const maxUserSessions = 64

type sessionStore struct{ db *bolt.DB }
type sessionRecord struct {
	UserID  uint
	Expires int64
	Proof   []byte
}

func sessionProof(id string) []byte  { hash := sha256.Sum256([]byte(id)); return hash[:] }
func sessionFamily(id string) string { family, _, _ := strings.Cut(id, "."); return family }
func sessionKey(id string) []byte    { return sessionProof(sessionFamily(id)) }

func (s sessionStore) Create(id string, userID uint, expires time.Time) error {
	if id == "" {
		return errors.New("empty session identifier")
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(sessionBucket)
		if err != nil {
			return err
		}
		// Expired sessions are collected on issuance, and per-user storage is
		// bounded even when a client repeatedly logs in without logging out.
		var owned [][]byte
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var r sessionRecord
			if err := json.Unmarshal(v, &r); err != nil {
				return err
			}
			if r.Expires <= time.Now().Unix() {
				if err := c.Delete(); err != nil {
					return err
				}
				continue
			}
			if r.UserID == userID {
				owned = append(owned, append([]byte(nil), k...))
			}
		}
		for len(owned) >= maxUserSessions {
			if err := b.Delete(owned[0]); err != nil {
				return err
			}
			owned = owned[1:]
		}
		value, err := json.Marshal(sessionRecord{UserID: userID, Expires: expires.Unix(), Proof: sessionProof(id)})
		if err != nil {
			return err
		}
		return b.Put(sessionKey(id), value)
	})
}

func (s sessionStore) Valid(id string, userID uint, now time.Time) (bool, error) {
	return s.valid(id, userID, now, true)
}
func (s sessionStore) ValidFamily(id string, userID uint, now time.Time) (bool, error) {
	return s.valid(id, userID, now, false)
}
func (s sessionStore) valid(id string, userID uint, now time.Time, current bool) (bool, error) {
	if id == "" {
		return false, nil
	}
	var record sessionRecord
	found := false
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionBucket)
		if b == nil {
			return nil
		}
		value := b.Get(sessionKey(id))
		if value == nil {
			return nil
		}
		found = true
		return json.Unmarshal(value, &record)
	})
	if err != nil || !found {
		return false, err
	}
	if record.Expires <= now.Unix() {
		return false, s.Delete(id)
	}
	return record.UserID == userID && (!current || hmac.Equal(record.Proof, sessionProof(id))), nil
}

func (s sessionStore) Delete(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionBucket)
		if b == nil {
			return nil
		}
		return b.Delete(sessionKey(id))
	})
}

// Replace consumes the old identifier and creates its successor in one write
// transaction. Two concurrent renewals cannot both redeem the same session.
func (s sessionStore) Replace(oldID, newID string, userID uint, expires time.Time) error {
	if oldID == "" || newID == "" || oldID == newID || sessionFamily(oldID) != sessionFamily(newID) {
		return storage.ErrInvalidSession
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionBucket)
		if b == nil {
			return storage.ErrInvalidSession
		}
		value := b.Get(sessionKey(oldID))
		if value == nil {
			return storage.ErrInvalidSession
		}
		var previous sessionRecord
		if err := json.Unmarshal(value, &previous); err != nil {
			return err
		}
		if previous.UserID != userID || previous.Expires <= time.Now().Unix() || !hmac.Equal(previous.Proof, sessionProof(oldID)) {
			return storage.ErrInvalidSession
		}
		next, err := json.Marshal(sessionRecord{UserID: userID, Expires: expires.Unix(), Proof: sessionProof(newID)})
		if err != nil {
			return err
		}
		if err := b.Delete(sessionKey(oldID)); err != nil {
			return err
		}
		return b.Put(sessionKey(newID), next)
	})
}
