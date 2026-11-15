package fbhttp

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/share"
	"github.com/filebrowser/filebrowser/v2/users"
)

// GHSA-7w29-q235-57m9: both the alias and its target must pass rules, including
// write destinations that do not yet exist below a symlinked directory.
func TestSymlinkAliasesEnforceTargetRules(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "denied"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "denied", "secret.txt"), []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}
	for alias, dest := range map[string]string{"alias": "denied", "dangling": "denied/new.txt"} {
		if err := os.Symlink(dest, filepath.Join(root, alias)); err != nil {
			t.Skip(err)
		}
	}
	key := []byte("test-signing-key")
	perm := users.Permissions{Download: true, Modify: true, Create: true}
	st := denyRuleStorage(t, root, "/denied", perm, key)
	for _, tc := range []struct {
		method, path string
		handler      handleFunc
	}{
		{"GET", "/alias/secret.txt", rawHandler},
		{"PUT", "/alias/secret.txt", resourcePutHandler},
		{"PUT", "/alias/new.txt", resourcePutHandler},
		{"PUT", "/dangling", resourcePutHandler},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader("changed"))
		req.Header.Set("X-Auth", signToken(t, perm, key, st))
		rec := httptest.NewRecorder()
		handle(tc.handler, "", st, &settings.Server{}).ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s %s: got %d want 403", tc.method, tc.path, rec.Code)
		}
	}
	data, err := os.ReadFile(filepath.Join(root, "denied", "secret.txt"))
	if err != nil || string(data) != "secret" {
		t.Fatalf("target changed: %q %v", data, err)
	}
	if _, err := os.Stat(filepath.Join(root, "denied", "new.txt")); !os.IsNotExist(err) {
		t.Fatal("denied target created")
	}
	// An alias to an allowed file remains usable.
	if err := os.WriteFile(filepath.Join(root, "allowed.txt"), []byte("ok"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("allowed.txt", filepath.Join(root, "allowed-link")); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/allowed-link", nil)
	req.Header.Set("X-Auth", signToken(t, perm, key, st))
	rec := httptest.NewRecorder()
	handle(rawHandler, "", st, &settings.Server{}).ServeHTTP(rec, req)
	if rec.Code != 200 || rec.Body.String() != "ok" {
		t.Fatalf("allowed alias: %d %s", rec.Code, rec.Body)
	}
}

func TestPublicShareSymlinkAliasesEnforceTargetRules(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "shared", "denied"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "shared", "denied", "secret.txt"), []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("denied", filepath.Join(root, "shared", "alias")); err != nil {
		t.Skip(err)
	}
	st := denyRuleStorage(t, root, "/shared/denied", users.Permissions{Download: true, Share: true}, []byte("key"))
	if err := st.Share.Save(&share.Link{Hash: "h", UserID: 1, Path: "/shared"}); err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "h/alias/secret.txt", nil)
	if err != nil {
		t.Fatal(err)
	}
	handle(publicDlHandler, "", st, &settings.Server{}).ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Fatalf("shared alias returned %d: %s", rec.Code, rec.Body)
	}
}
