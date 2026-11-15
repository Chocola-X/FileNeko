package fbhttp

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
)

func TestWebDAVLockTokensAreNotSequential(t *testing.T) {
	root := t.TempDir()
	h, session, _ := davSecurityFixture(t, root)
	body := `<D:lockinfo xmlns:D="DAV:"><D:lockscope><D:exclusive/></D:lockscope><D:locktype><D:write/></D:locktype></D:lockinfo>`
	w := davSecurityRequest(h, session, "LOCK", "/new.txt", body, map[string]string{"Depth": "0"})
	if w.Code != 201 {
		t.Fatalf("lock status=%d: %s", w.Code, w.Body)
	}
	token := w.Header().Get("Lock-Token")
	if len(token) < 32 {
		t.Fatalf("lock token lacks unpredictable entropy: %q", token)
	}
	if _, err := strconv.ParseUint(token[1:len(token)-1], 10, 64); err == nil {
		t.Fatal("lock tokens expose a sequential timestamp counter")
	}
}

func TestWebDAVLockRegistryExpiresAndRejectsInternalTokens(t *testing.T) {
	system := newDAVLockRegistry()
	now := time.Now()
	details := webdav.LockDetails{Root: "/scope/file", Duration: time.Minute, ZeroDepth: true}
	token, err := system.Create(now, details, 1, "/file")
	if err != nil {
		t.Fatal(err)
	}
	internal := system.entries[token].internal
	if _, err := system.Confirm(now, details.Root, "", webdav.Condition{Token: internal}); !errors.Is(err, webdav.ErrConfirmationFailed) {
		t.Fatal("internal sequential token reached the underlying lock system")
	}
	release, err := system.Confirm(now, details.Root, "", webdav.Condition{Token: token})
	if err != nil {
		t.Fatal(err)
	}
	release()
	if _, err := system.Create(now.Add(time.Minute), details, 1, "/file"); err != nil {
		t.Fatalf("expired lock still prevents a new lock: %v", err)
	}
	if _, ok := system.entries[token]; ok {
		t.Fatal("expired capability was not collected")
	}
	for len(system.entries) < maxDAVLocks {
		system.entries[strconv.Itoa(len(system.entries))] = davLockEntry{}
	}
	if _, err := system.Create(now.Add(time.Minute), webdav.LockDetails{Root: "/other"}, 1, "/other"); !errors.Is(err, webdav.ErrLocked) {
		t.Fatal("lock tracking is unbounded")
	}
}

func TestWebDAVCannotUnlockOutsideScopeOrDeniedRules(t *testing.T) {
	root := t.TempDir()
	makeData := func(name string) *data {
		t.Helper()
		dir := filepath.Join(root, name)
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
		return &data{settings: &settings.Settings{}, server: &settings.Server{}, user: &users.User{Fs: files.NewFs(afero.NewOsFs(), dir, false)}}
	}
	system := newDAVLockRegistry()
	alice := davLocks{locks: system, d: makeData("alice")}
	bob := davLocks{locks: system, d: makeData("bob")}
	now := time.Now()
	token, err := alice.Create(now, webdav.LockDetails{Root: "/private.txt", Duration: time.Minute, ZeroDepth: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := bob.Unlock(now, token); err == nil {
		t.Fatal("other scope can unlock a resource it cannot access")
	}
	if _, err := bob.Refresh(now, token, time.Hour); err == nil {
		t.Fatal("other scope can refresh a resource it cannot access")
	}
	if !system.entries[token].expires.Equal(now.Add(time.Minute)) {
		t.Fatal("rejected refresh mutated the expiry before checking authorization")
	}
	alice.d.settings.Rules = []rules.Rule{{Path: "/private.txt", Allow: false}}
	if err := alice.Unlock(now, token); err == nil {
		t.Fatal("deny rule did not protect locked resource")
	}
	alice.d.settings.Rules = nil
	if err := alice.Unlock(now, token); err != nil {
		t.Fatal(err)
	}
}
