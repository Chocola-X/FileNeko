package fbhttp

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/asdine/storm/v3"
	"github.com/filebrowser/filebrowser/v2/diskcache"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/share"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
	"github.com/filebrowser/filebrowser/v2/users"
)

// GHSA-r6pg-pg54-rcr5 and the negative control from GHSA-5ww9-jg6q-38r7:
// revoke different owners' links to the same object, retain same-name links in
// other scopes, and account for nested scopes using different virtual paths.
func TestDeleteRevokesSharesByPhysicalPath(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"common/nested", "other"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"common/nested/a.txt", "other/a.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("old"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	db, err := storm.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	st, err := bolt.NewStorage(db)
	if err != nil {
		t.Fatal(err)
	}
	perm := users.Permissions{Delete: true, Download: true, Share: true}
	for idx, scope := range []string{"common", "common", "common/nested", "other"} {
		if err := st.Users.Save(&users.User{Username: string(rune('a' + idx)), Password: "pw", Scope: scope, Perm: perm}); err != nil {
			t.Fatal(err)
		}
	}
	key := []byte("key")
	if err := st.Settings.Save(&settings.Settings{Key: key}); err != nil {
		t.Fatal(err)
	}
	for _, link := range []*share.Link{
		{Hash: "own", UserID: 1, Path: "/nested/a.txt"},
		{Hash: "same", UserID: 2, Path: "/nested/a.txt"},
		{Hash: "nested", UserID: 3, Path: "/a.txt"},
		{Hash: "other", UserID: 4, Path: "/a.txt"},
		{Hash: "prefix", UserID: 2, Path: "/nested/a.txt.other"},
	} {
		if err := st.Share.Save(link); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest("DELETE", "/nested/a.txt", nil)
	req.Header.Set("X-Auth", signToken(t, perm, key))
	rec := httptest.NewRecorder()
	handle(resourceDeleteHandler(diskcache.NewNoOp()), "", st, &settings.Server{Root: root}).ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Fatalf("delete: %d %s", rec.Code, rec.Body)
	}
	for _, hash := range []string{"own", "same", "nested"} {
		if _, err := st.Share.GetByHash(hash); err == nil {
			t.Errorf("share %s survived", hash)
		}
	}
	for _, hash := range []string{"other", "prefix"} {
		if _, err := st.Share.GetByHash(hash); err != nil {
			t.Errorf("unrelated share %s removed: %v", hash, err)
		}
	}
}
