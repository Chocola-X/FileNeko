package auth

import (
	"errors"
	"github.com/filebrowser/filebrowser/v2/settings"
	"sync"
	"testing"
	"time"
)

func TestPasswordGuessBudgetAndRecovery(t *testing.T) {
	l := newPasswordLimiter()
	now := time.Now()
	l.now = func() time.Time { return now }
	policy := settings.DefaultFail2Ban()
	policy.Enabled = false
	for i := 0; i < passwordBurst; i++ {
		finish, err := l.begin("192.0.2.1:1234", policy)
		if err != nil {
			t.Fatalf("attempt %d rejected early", i)
		}
		finish(false)
	}
	if finish, err := l.begin("192.0.2.1:5678", policy); err == nil {
		finish(false)
		t.Fatal("changing source port bypassed limit")
	}
	finish, err := l.begin("192.0.2.2:1234", policy)
	if err != nil {
		t.Fatal("independent peer blocked")
	}
	finish(true)
	now = now.Add(passwordRefill)
	finish, err = l.begin("192.0.2.1:1234", policy)
	if err != nil {
		t.Fatal("budget did not refill")
	}
	finish(true)
}
func TestPasswordChecksAndMemoryAreBounded(t *testing.T) {
	l := newPasswordLimiter()
	now := time.Now()
	l.now = func() time.Time { return now }
	policy := settings.DefaultFail2Ban()
	policy.Enabled = false
	var releases []func(bool)
	for i := 0; i < cap(l.workers); i++ {
		finish, err := l.begin("192.0.2.1:1", policy)
		if err != nil {
			t.Fatal("worker missing")
		}
		releases = append(releases, finish)
	}
	if finish, err := l.begin("192.0.2.2:1", policy); err == nil {
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
	if finish, err := l.begin("new-peer", policy); err == nil {
		finish(false)
		t.Fatal("unbounded peer tracking")
	}
	now = now.Add(passwordPeerTTL + time.Minute)
	finish, err := l.begin("new-peer", policy)
	if err != nil {
		t.Fatal("expired peers not reclaimed")
	}
	finish(true)
}

func TestFail2BanWindowExpiryAndPolicyChanges(t *testing.T) {
	l := newPasswordLimiter()
	now := time.Now()
	l.now = func() time.Time { return now }
	policy := settings.Fail2Ban{Enabled: true, MaxAttempts: 3, WindowMinutes: 2, BanMinutes: 5}
	attempt := func(success bool) {
		t.Helper()
		finish, err := l.begin("192.0.2.3:1234", policy)
		if err != nil {
			t.Fatalf("attempt blocked: %v", err)
		}
		finish(success)
	}
	attempt(false)
	now = now.Add(2 * time.Minute)
	attempt(false) // The old observation window has expired.
	attempt(true)  // Successful DAV traffic does not count as a failure.
	attempt(false)
	if p := l.peers["192.0.2.3"]; p.failures != 2 {
		t.Fatalf("failures=%d", p.failures)
	}
	attempt(false)
	deadline := now.Add(5 * time.Minute)
	for _, advance := range []time.Duration{0, time.Minute, 3 * time.Minute} {
		now = now.Add(advance)
		_, err := l.begin("192.0.2.3:9876", policy)
		var limit *AttemptLimitError
		if !errors.As(err, &limit) || limit.RetryAfter != deadline.Sub(now) {
			t.Fatalf("missing ban or incorrect remaining duration: %v", err)
		}
	}
	now = now.Add(time.Minute)
	attempt(true)
	attempt(false)
	attempt(false)
	attempt(false)
	// Disabling the configurable ban immediately removes it; basic work limits remain.
	policy.Enabled = false
	attempt(true)
}

func TestFail2BanCountsIPv6PrefixAndConcurrentFailures(t *testing.T) {
	l := newPasswordLimiter()
	now := time.Now()
	l.now = func() time.Time { return now }
	policy := settings.DefaultFail2Ban()
	var finishes []func(bool)
	for i := 0; i < 5; i++ {
		finish, err := l.begin("[2001:db8:abcd:1::1]:42", policy)
		if err != nil {
			t.Fatal(err)
		}
		finishes = append(finishes, finish)
	}
	var wg sync.WaitGroup
	for _, finish := range finishes {
		wg.Add(1)
		go func() { defer wg.Done(); finish(false) }()
	}
	wg.Wait()
	for _, addr := range []string{"[2001:db8:abcd:1::2]:99", "[2001:db8:abcd:1::ffff]:1234"} {
		if finish, err := l.begin(addr, policy); err == nil {
			finish(true)
			t.Fatal("IPv6 suffix bypassed ban")
		}
	}
	finish, err := l.begin("[2001:db8:abcd:2::1]:42", policy)
	if err != nil {
		t.Fatal("independent IPv6 prefix blocked")
	}
	finish(true)
}

func TestFail2BanDoesNotExtendOnBlockedRequestsOrForgetLongBans(t *testing.T) {
	l := newPasswordLimiter()
	now := time.Now()
	l.now = func() time.Time { return now }
	policy := settings.Fail2Ban{Enabled: true, MaxAttempts: 1, WindowMinutes: 1, BanMinutes: 60}
	finish, err := l.begin("192.0.2.4:1", policy)
	if err != nil {
		t.Fatal(err)
	}
	finish(false)
	deadline := l.peers["192.0.2.4"].bannedUntil
	now = now.Add(11 * time.Minute)
	// Force a cache sweep after the normal idle TTL; an active ban must survive.
	for len(l.peers) < maxPasswordPeers {
		l.peers[string(rune(len(l.peers)))] = &passwordPeer{seen: now, updated: now}
	}
	if _, err := l.begin("192.0.2.4:2", policy); err == nil {
		t.Fatal("long ban expired early")
	}
	if l.peers["192.0.2.4"].bannedUntil != deadline {
		t.Fatal("rejected request extended ban")
	}
}
