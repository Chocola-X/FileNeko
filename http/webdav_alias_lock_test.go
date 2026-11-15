package fbhttp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWebDAVLocksCoverSymlinkAliases(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "original.txt"), []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("original.txt", filepath.Join(root, "alias.txt")); err != nil {
		t.Skip(err)
	}
	h, session, _ := davSecurityFixture(t, root)
	body := `<D:lockinfo xmlns:D="DAV:"><D:lockscope><D:exclusive/></D:lockscope><D:locktype><D:write/></D:locktype></D:lockinfo>`
	lock := davSecurityRequest(h, session, "LOCK", "/original.txt", body, map[string]string{"Depth": "0"})
	if lock.Code != 200 {
		t.Fatalf("lock status=%d: %s", lock.Code, lock.Body)
	}
	token := lock.Header().Get("Lock-Token")
	if w := davSecurityRequest(h, session, "PUT", "/alias.txt", "unauthorized", nil); w.Code != 423 {
		t.Fatalf("symlink bypassed lock: status=%d", w.Code)
	}
	if data, err := os.ReadFile(filepath.Join(root, "original.txt")); err != nil || string(data) != "original" {
		t.Fatal("locked target changed")
	}
	if w := davSecurityRequest(h, session, "PUT", "/alias.txt", "authorized", map[string]string{"If": "(" + token + ")"}); w.Code != 201 {
		t.Fatalf("authorized alias write rejected: %d %s", w.Code, w.Body)
	}
	if w := davSecurityRequest(h, session, "LOCK", "/alias.txt", "", map[string]string{"If": "(" + token + ")"}); w.Code != 200 {
		t.Fatalf("alias lock refresh rejected: %d", w.Code)
	}
	if w := davSecurityRequest(h, session, "UNLOCK", "/original.txt", "", map[string]string{"Lock-Token": token}); w.Code != 204 {
		t.Fatalf("unlock status=%d", w.Code)
	}
}
