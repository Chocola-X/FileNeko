package fbhttp

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/filebrowser/filebrowser/v2/diskcache"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// GHSA-c4fr-5f24-4wrj: an upload must never delete an existing directory on
// failure, including descendants protected by a deny rule.
func TestUploadOverrideDoesNotDeleteDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "team"), 0o755); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(root, "team", "secret.txt")
	if err := os.WriteFile(secret, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	key := []byte("test-signing-key")
	perm := users.Permissions{Create: true, Modify: true}
	st := denyRuleStorage(t, root, "/team/secret.txt", perm, key)
	req := httptest.NewRequest(http.MethodPost, "/team?override=true", strings.NewReader("replace"))
	req.Header.Set("X-Auth", signToken(t, perm, key, st))
	rec := httptest.NewRecorder()
	handle(resourcePostHandler(diskcache.NewNoOp()), "", st, &settings.Server{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body)
	}
	if got, err := os.ReadFile(secret); err != nil || string(got) != "keep" {
		t.Fatalf("protected descendant changed: %q, %v", got, err)
	}
}
