package bolt

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

func TestSessionsPersistAndRevoke(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.db")
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	s := sessionStore{db: db}
	expiry := time.Now().Add(time.Hour)
	if err := s.Create("random-session", 1, expiry); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s = sessionStore{db: db}
	if ok, err := s.Valid("random-session", 1, time.Now()); err != nil || !ok {
		t.Fatalf("persisted session: %v %v", ok, err)
	}
	if ok, err := s.Valid("random-session", 2, time.Now()); err != nil || ok {
		t.Fatalf("wrong user: %v %v", ok, err)
	}
	if err := s.Delete("random-session"); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.Valid("random-session", 1, time.Now()); err != nil || ok {
		t.Fatalf("revoked session: %v %v", ok, err)
	}
}

func TestSessionsExpireAndStayBounded(t *testing.T) {
	db, err := bolt.Open(filepath.Join(t.TempDir(), "sessions.db"), 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := sessionStore{db: db}
	if err := s.Create("expired", 1, time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.Valid("expired", 1, time.Now()); err != nil || ok {
		t.Fatalf("expired: %v %v", ok, err)
	}
	for i := 0; i < maxUserSessions+10; i++ {
		if err := s.Create(fmt.Sprint(i), 1, time.Now().Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.View(func(tx *bolt.Tx) error {
		if n := tx.Bucket(sessionBucket).Stats().KeyN; n != maxUserSessions {
			t.Errorf("stored %d sessions, want %d", n, maxUserSessions)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
