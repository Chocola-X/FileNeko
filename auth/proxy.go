package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"os"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// MethodProxyAuth is used to identify no auth.
const MethodProxyAuth settings.AuthMethod = "proxy"

// ProxyAuth is a proxy implementation of an auther.
type ProxyAuth struct {
	Header         string   `json:"header"`
	TrustedProxies []string `json:"trustedProxies"`
}

// Auth authenticates the user via an HTTP header.
func (a ProxyAuth) Auth(r *http.Request, usr users.Store, setting *settings.Settings, srv *settings.Server) (*users.User, error) {
	if !a.TrustedPeer(r.RemoteAddr) {
		return nil, os.ErrPermission
	}
	username := r.Header.Get(a.Header)
	if a.Header == "" || username == "" {
		return nil, os.ErrPermission
	}
	user, err := usr.Get(srv.Root, srv.FollowExternalSymlinks, username)
	if errors.Is(err, fberrors.ErrNotExist) {
		return a.createUser(usr, setting, srv, username)
	}
	return user, err
}

func (a ProxyAuth) createUser(usr users.Store, setting *settings.Settings, srv *settings.Server, username string) (*users.User, error) {
	const randomPasswordLength = settings.DefaultMinimumPasswordLength + 10
	pwd, err := users.RandomPwd(randomPasswordLength)
	if err != nil {
		return nil, err
	}

	var hashedRandomPassword string
	hashedRandomPassword, err = users.ValidateAndHashPwd(pwd, setting.MinimumPasswordLength)
	if err != nil {
		return nil, err
	}

	user := &users.User{
		Username:     username,
		Password:     hashedRandomPassword,
		LockPassword: true,
	}
	setting.Defaults.Apply(user)
	user.Perm.Admin = false
	user.Perm.Execute = false
	user.Commands = []string{}

	var derivedScope bool
	if derivedScope, err = setting.CreateUserHome(user, srv.Root, false); err != nil {
		return nil, err
	}

	if err = usr.SaveProvisioned(user, derivedScope); err != nil {
		return nil, err
	}

	return user, nil
}

// LoginPage tells that proxy auth doesn't require a login page.
func (a ProxyAuth) LoginPage() bool {
	return false
}

// TrustedPeer uses only the connection peer, never attacker-supplied forwarded
// headers. Existing installations default to loopback; remote proxies must be
// explicitly configured as individual IP addresses or CIDR ranges.
func (a ProxyAuth) TrustedPeer(remote string) bool {
	peer, err := netip.ParseAddrPort(remote)
	if err != nil {
		return false
	}
	addr := peer.Addr().Unmap()
	if len(a.TrustedProxies) == 0 {
		return addr.IsLoopback()
	}
	for _, entry := range a.TrustedProxies {
		if ip, err := netip.ParseAddr(entry); err == nil && ip.Unmap() == addr {
			return true
		}
		if prefix, err := netip.ParsePrefix(entry); err == nil && prefix.Contains(addr) {
			return true
		}
	}
	return false
}

func (a ProxyAuth) ValidateTrustedProxies() error {
	for _, entry := range a.TrustedProxies {
		if _, err := netip.ParseAddr(entry); err == nil {
			continue
		}
		if _, err := netip.ParsePrefix(entry); err != nil {
			return fmt.Errorf("invalid trusted proxy IP or CIDR: %q", entry)
		}
	}
	return nil
}
