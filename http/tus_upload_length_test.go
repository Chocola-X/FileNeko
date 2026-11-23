package fbhttp

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

func TestTusEmptyUploadSurvivesCacheExpiry(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		checkTusEmptyUploadSurvivesCacheExpiry(t, func() {
			time.Sleep(uploadCacheTTL + time.Second)
			synctest.Wait()
		})
	})
}

func checkTusEmptyUploadSurvivesCacheExpiry(t *testing.T, waitForExpiry func()) {
	t.Helper()
	scope := t.TempDir()
	key := []byte("test-signing-key")
	perm := users.Permissions{Create: true, Modify: true, Download: true}
	st := scopedUserStorage(t, scope, perm, key)
	token := signToken(t, perm, key, st)

	cache := newMemoryUploadCache()
	t.Cleanup(cache.Close)
	handler, err := NewHandler(nil, nil, cache, st, &settings.Server{
		Root: scope, EnableWebDAV: true,
	}, os.DirFS(scope))
	if err != nil {
		t.Fatal(err)
	}

	request := func(method, path, length string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("X-Auth", token)
		if method == http.MethodPost {
			req.Header.Set("Upload-Length", length)
		}
		if method == "PROPFIND" {
			req.Header.Set("Depth", "1")
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}
	post := func(path, length, location string) {
		t.Helper()
		res := request(http.MethodPost, path, length)
		if res.Code != http.StatusCreated || res.Header().Get("Location") != location {
			t.Fatalf("POST %s: got status %d, Location %q, body %q", path, res.Code, res.Header().Get("Location"), res.Body.String())
		}
	}

	post("/api/tus/abandoned.txt", "5", "/api/tus/abandoned.txt")
	post("/api/tus/overwrite.txt", "5", "/api/tus/overwrite.txt")
	post("/api/tus/overwrite.txt?override=true", "0", "/api/tus/overwrite.txt")
	post("/api/tus/empty.txt", "0", "/api/tus/empty.txt")

	waitForExpiry()
	if _, err := os.Stat(filepath.Join(scope, "abandoned.txt")); !os.IsNotExist(err) {
		t.Errorf("abandoned upload should be removed after expiry, got %v", err)
	}
	api := request(http.MethodGet, "/api/resources/", "")
	dav := request("PROPFIND", "/dav/", "")
	for _, name := range []string{"empty.txt", "overwrite.txt"} {
		info, err := os.Stat(filepath.Join(scope, name))
		if err != nil {
			t.Errorf("stat %s after expiry: %v", name, err)
		} else if info.Size() != 0 {
			t.Errorf("%s size after expiry = %d, want 0", name, info.Size())
		}
		if api.Code != http.StatusOK || !strings.Contains(api.Body.String(), `"name":"`+name+`"`) {
			t.Errorf("API listing should include %s, got %d: %s", name, api.Code, api.Body.String())
		}
		if dav.Code != http.StatusMultiStatus || !strings.Contains(dav.Body.String(), "/dav/"+name+"</") {
			t.Errorf("WebDAV listing should include %s, got %d: %s", name, dav.Code, dav.Body.String())
		}
	}
}

// A TUS PATCH must not write more than the declared Upload-Length. A client that
// declares a small length and then streams a large body must be rejected without
// the excess being written to disk, while a correctly-sized upload still works.
func TestTusPatchEnforcesUploadLength(t *testing.T) {
	root := t.TempDir()
	userScope := filepath.Join(root, "user")
	if err := os.MkdirAll(userScope, 0o755); err != nil {
		t.Fatal(err)
	}

	key := []byte("test-signing-key")
	perm := users.Permissions{Create: true, Modify: true}
	st := scopedUserStorage(t, userScope, perm, key)
	signed := signToken(t, perm, key, st)

	cache := newMemoryUploadCache()
	t.Cleanup(cache.Close)
	post := handle(tusPostHandler(cache), "", st, &settings.Server{})
	patch := handle(tusPatchHandler(cache), "", st, &settings.Server{})

	patchReq := func(body string) *httptest.ResponseRecorder {
		req, _ := http.NewRequest(http.MethodPatch, "/file.txt", strings.NewReader(body))
		req.Header.Set("X-Auth", signed)
		req.Header.Set("Content-Type", "application/offset+octet-stream")
		req.Header.Set("Upload-Offset", "0")
		rec := httptest.NewRecorder()
		patch.ServeHTTP(rec, req)
		return rec
	}

	// Create an upload that declares only 5 bytes.
	reqPost, _ := http.NewRequest(http.MethodPost, "/file.txt", http.NoBody)
	reqPost.Header.Set("X-Auth", signed)
	reqPost.Header.Set("Upload-Length", "5")
	recPost := httptest.NewRecorder()
	post.ServeHTTP(recPost, reqPost)
	if recPost.Code != http.StatusCreated {
		t.Fatalf("POST expected 201, got %d body=%q", recPost.Code, recPost.Body.String())
	}

	// A body far larger than the declared length must be rejected, and nothing
	// beyond the declared length may reach disk.
	if rec := patchReq(strings.Repeat("A", 5000)); rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("over-length PATCH expected 413, got %d body=%q", rec.Code, rec.Body.String())
	}
	if fi, err := os.Stat(filepath.Join(userScope, "file.txt")); err != nil {
		t.Fatalf("stat file.txt: %v", err)
	} else if fi.Size() > 5 {
		t.Fatalf("wrote %d bytes despite Upload-Length 5", fi.Size())
	}

	// A correctly-sized upload still completes.
	if rec := patchReq("hello"); rec.Code != http.StatusNoContent {
		t.Fatalf("valid PATCH expected 204, got %d body=%q", rec.Code, rec.Body.String())
	}
	if data, _ := os.ReadFile(filepath.Join(userScope, "file.txt")); string(data) != "hello" {
		t.Fatalf("expected file content \"hello\", got %q", string(data))
	}
}
