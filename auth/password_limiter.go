package auth

import (
	"net"
	"net/netip"
	"sync"
	"time"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
)

const (
	passwordBurst    = 10
	passwordRefill   = 6 * time.Second
	passwordPeerTTL  = 10 * time.Minute
	maxPasswordPeers = 4096
)

// AttemptLimitError reports how long a caller should wait without exposing
// whether an account exists. It also covers the independent bcrypt work limit.
type AttemptLimitError struct{ RetryAfter time.Duration }

func (e *AttemptLimitError) Error() string { return fberrors.ErrTooManyAttempts.Error() }
func (e *AttemptLimitError) Unwrap() error { return fberrors.ErrTooManyAttempts }

type passwordPeer struct {
	credits                  float64
	updated, seen            time.Time
	policy                   settings.Fail2Ban
	failures                 uint
	windowStart, bannedUntil time.Time
	inflight                 uint
}
type passwordLimiter struct {
	mu      sync.Mutex
	peers   map[string]*passwordPeer
	workers chan struct{}
	swept   time.Time
	now     func() time.Time
}

func newPasswordLimiter() *passwordLimiter {
	return &passwordLimiter{peers: make(map[string]*passwordPeer), workers: make(chan struct{}, 8), now: time.Now}
}

var passwordAttempts = newPasswordLimiter()

func passwordPeerKey(remote string) string {
	peer, _, err := net.SplitHostPort(remote)
	if err != nil {
		peer = remote
	}
	if ip, err := netip.ParseAddr(peer); err == nil {
		ip = ip.Unmap()
		if ip.Is6() {
			return netip.PrefixFrom(ip, 64).Masked().String()
		}
		return ip.String()
	}
	return peer
}

// begin reserves a password check. Successful requests refund their work credit
// but do not erase other failed attempts in the observation window. Only the
// actual connection source is trusted; forwarded headers cannot mint new keys.
func (l *passwordLimiter) begin(remote string, policy settings.Fail2Ban) (func(bool), error) {
	now := l.now()
	peer := passwordPeerKey(remote)
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.peers) >= maxPasswordPeers && now.Sub(l.swept) > time.Minute {
		for key, p := range l.peers {
			retention := max(passwordPeerTTL, time.Duration(p.policy.WindowMinutes)*time.Minute)
			if p.inflight == 0 && !now.Before(p.bannedUntil) && now.Sub(p.seen) > retention {
				delete(l.peers, key)
			}
		}
		l.swept = now
	}
	p := l.peers[peer]
	if p == nil {
		if len(l.peers) >= maxPasswordPeers {
			return nil, &AttemptLimitError{passwordRefill}
		}
		p = &passwordPeer{credits: passwordBurst, updated: now, policy: policy}
		l.peers[peer] = p
	}
	p.seen = now
	if p.policy != policy {
		p.policy = policy
		p.failures = 0
		p.windowStart = time.Time{}
		// Turning protection off takes effect immediately. Changing limits does
		// not shorten a ban already imposed under the previous enabled policy.
		if !policy.Enabled {
			p.bannedUntil = time.Time{}
		}
	}
	if policy.Enabled && now.Before(p.bannedUntil) {
		return nil, &AttemptLimitError{p.bannedUntil.Sub(now)}
	}
	if !p.bannedUntil.IsZero() {
		p.bannedUntil = time.Time{}
		p.failures = 0
		p.windowStart = time.Time{}
	}
	p.credits = min(passwordBurst, p.credits+now.Sub(p.updated).Seconds()/passwordRefill.Seconds())
	p.updated = now
	if p.credits < 1 {
		return nil, &AttemptLimitError{time.Duration((1 - p.credits) * float64(passwordRefill))}
	}
	select {
	case l.workers <- struct{}{}:
	default:
		return nil, &AttemptLimitError{passwordRefill}
	}
	p.credits--
	p.inflight++
	return func(success bool) {
		finished := l.now()
		l.mu.Lock()
		p.inflight--
		if success {
			p.credits = min(passwordBurst, p.credits+1)
		} else if policy.Enabled && p.policy == policy && !finished.Before(p.bannedUntil) {
			if p.windowStart.IsZero() || finished.Sub(p.windowStart) >= time.Duration(policy.WindowMinutes)*time.Minute {
				p.windowStart = finished
				p.failures = 0
			}
			p.failures++
			if p.failures >= policy.MaxAttempts {
				p.bannedUntil = finished.Add(time.Duration(policy.BanMinutes) * time.Minute)
			}
		}
		l.mu.Unlock()
		<-l.workers
	}, nil
}
