package auth

import (
	"net"
	"net/netip"
	"sync"
	"time"
)

const (
	passwordBurst    = 10
	passwordRefill   = 6 * time.Second
	passwordPeerTTL  = 10 * time.Minute
	maxPasswordPeers = 4096
)

type passwordPeer struct {
	credits       float64
	updated, seen time.Time
}
type passwordLimiter struct {
	mu      sync.Mutex
	peers   map[string]*passwordPeer
	workers chan struct{}
	swept   time.Time
}

func newPasswordLimiter() *passwordLimiter {
	return &passwordLimiter{peers: make(map[string]*passwordPeer), workers: make(chan struct{}, 8)}
}

var passwordAttempts = newPasswordLimiter()

// begin reserves a password check. Successful requests refund their credit so
// normal WebDAV traffic does not consume the failed-login budget. Trust only
// the connection address: forwarded headers cannot mint fresh rate-limit keys.
func (l *passwordLimiter) begin(remote string, now time.Time) (func(bool), bool) {
	select {
	case l.workers <- struct{}{}:
	default:
		return nil, false
	}
	peer, _, err := net.SplitHostPort(remote)
	if err != nil {
		peer = remote
	}

	if ip, err := netip.ParseAddr(peer); err == nil {
		ip = ip.Unmap()
		if ip.Is6() {
			peer = netip.PrefixFrom(ip, 64).Masked().String()
		} else {
			peer = ip.String()
		}
	}
	l.mu.Lock()
	if len(l.peers) >= maxPasswordPeers && now.Sub(l.swept) > time.Minute {
		for key, p := range l.peers {
			if now.Sub(p.seen) > passwordPeerTTL {
				delete(l.peers, key)
			}
		}
		l.swept = now
	}
	p := l.peers[peer]
	if p == nil {
		if len(l.peers) >= maxPasswordPeers {
			l.mu.Unlock()
			<-l.workers
			return nil, false
		}
		p = &passwordPeer{credits: passwordBurst, updated: now}
		l.peers[peer] = p
	}
	p.credits = min(passwordBurst, p.credits+now.Sub(p.updated).Seconds()/passwordRefill.Seconds())
	p.updated = now
	p.seen = now
	if p.credits < 1 {
		l.mu.Unlock()
		<-l.workers
		return nil, false
	}
	p.credits--
	l.mu.Unlock()
	return func(success bool) {
		if success {
			l.mu.Lock()
			p.credits = min(passwordBurst, p.credits+1)
			l.mu.Unlock()
		}
		<-l.workers
	}, true
}
