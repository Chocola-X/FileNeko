package fbhttp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWebDAVTaggedIfCannotAuthorizeAnotherResource(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"locked.txt", "other.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("original"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	h, session, _ := davSecurityFixture(t, root)
	body := `<D:lockinfo xmlns:D="DAV:"><D:lockscope><D:exclusive/></D:lockscope><D:locktype><D:write/></D:locktype></D:lockinfo>`
	for _, name := range []string{"locked.txt", "other.txt"} {
		w := davSecurityRequest(h, session, "LOCK", "/"+name, body, map[string]string{"Depth": "0"})
		if w.Code != 200 {
			t.Fatalf("LOCK %s: %d", name, w.Code)
		}
		if name == "other.txt" {
			token := w.Header().Get("Lock-Token")
			w = davSecurityRequest(h, session, "PUT", "/locked.txt", "bypassed", map[string]string{"If": "<http://example.com/dav/other.txt> (" + token + ")"})
			if w.Code != 412 {
				t.Fatalf("unrelated tagged resource authorized locked write: %d", w.Code)
			}
		}
	}
	content, err := os.ReadFile(filepath.Join(root, "locked.txt"))
	if err != nil || string(content) != "original" {
		t.Fatal("locked target was modified")
	}
}
