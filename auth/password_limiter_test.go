package auth

import (
	"testing"
	"time"
)

func TestPasswordGuessBudgetAndRecovery(t *testing.T) {
	l := newPasswordLimiter()
	now := time.Now()
	for i := 0; i < passwordBurst; i++ {
		finish, ok := l.begin("192.0.2.1:1234", now)
		if !ok {
			t.Fatalf("attempt %d rejected early", i)
		}
		finish(false)
	}
	if finish, ok := l.begin("192.0.2.1:5678", now); ok {
		finish(false)
		t.Fatal("changing source port bypassed limit")
	}
	finish, ok := l.begin("192.0.2.2:1234", now)
	if !ok {
		t.Fatal("independent peer blocked")
	}
	finish(true)
	finish, ok = l.begin("192.0.2.1:1234", now.Add(passwordRefill))
	if !ok {
		t.Fatal("budget did not refill")
	}
	finish(true)
}
func TestPasswordChecksAndMemoryAreBounded(t *testing.T) {
	l := newPasswordLimiter()
	now := time.Now()
	var releases []func(bool)
	for i := 0; i < cap(l.workers); i++ {
		finish, ok := l.begin("192.0.2.1:1", now)
		if !ok {
			t.Fatal("worker missing")
		}
		releases = append(releases, finish)
	}
	if finish, ok := l.begin("192.0.2.2:1", now); ok {
		finish(false)
		t.Fatal("unbounded concurrent password checks")
	}
	for _, finish := range releases {
		finish(true)
	}
	l.peers = make(map[string]*passwordPeer)
	for i := 0; i < maxPasswordPeers; i++ {
		l.peers[string(rune(i))] = &passwordPeer{credits: 1, updated: now, seen: now}
	}
	if finish, ok := l.begin("new-peer", now); ok {
		finish(false)
		t.Fatal("unbounded peer tracking")
	}
	finish, ok := l.begin("new-peer", now.Add(passwordPeerTTL+time.Minute))
	if !ok {
		t.Fatal("expired peers not reclaimed")
	}
	finish(true)
}
