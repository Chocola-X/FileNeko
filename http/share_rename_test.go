package fbhttp

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/filebrowser/filebrowser/v2/diskcache"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/share"
	"github.com/filebrowser/filebrowser/v2/users"
)

// GHSA-m8v4-4w34-rrvf: recreating a renamed path must not resurrect old links.
func TestRenameRevokesShares(t *testing.T) {
	for _, src := range []string{"/a.txt", "/dir"} {
		t.Run(src, func(t *testing.T) {
			root := t.TempDir()
			if err := os.MkdirAll(filepath.Join(root, "dir"), 0755); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"a.txt", "dir/child.txt"} {
				if err := os.WriteFile(filepath.Join(root, name), []byte("old"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			key := []byte("key")
			perm := users.Permissions{Rename: true, Share: true, Download: true}
			st := scopedUserStorage(t, root, perm, key)
			for hash, path := range map[string]string{"root": src, "child": "/dir/child.txt", "unrelated": "/a.txt.other"} {
				if err := st.Share.Save(&share.Link{Hash: hash, Path: path, UserID: 1}); err != nil {
					t.Fatal(err)
				}
			}
			req := httptest.NewRequest("PATCH", src+"?action=rename&destination=/new", nil)
			req.Header.Set("X-Auth", signToken(t, perm, key))
			rec := httptest.NewRecorder()
			handle(resourcePatchHandler(diskcache.NewNoOp()), "", st, &settings.Server{}).ServeHTTP(rec, req)
			if rec.Code >= 400 {
				t.Fatalf("rename: %d %s", rec.Code, rec.Body)
			}
			if _, err := st.Share.GetByHash("root"); err == nil {
				t.Fatal("old share survived rename")
			}
			if src == "/dir" {
				if _, err := st.Share.GetByHash("child"); err == nil {
					t.Fatal("descendant share survived rename")
				}
			}
			if _, err := st.Share.GetByHash("unrelated"); err != nil {
				t.Fatal("unrelated prefix share removed")
			}
		})
	}
}
