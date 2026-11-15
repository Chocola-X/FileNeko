package bolt

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"time"

	bolt "go.etcd.io/bbolt"
)

var sessionBucket = []byte("sessions")

const maxUserSessions = 64

type sessionStore struct{ db *bolt.DB }
type sessionRecord struct {
	UserID  uint
	Expires int64
}

func sessionKey(id string) []byte { hash := sha256.Sum256([]byte(id)); return hash[:] }

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
		value, err := json.Marshal(sessionRecord{UserID: userID, Expires: expires.Unix()})
		if err != nil {
			return err
		}
		return b.Put(sessionKey(id), value)
	})
}

func (s sessionStore) Valid(id string, userID uint, now time.Time) (bool, error) {
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
	return record.UserID == userID, nil
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
