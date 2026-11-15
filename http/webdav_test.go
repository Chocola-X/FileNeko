package fbhttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	fbAuth "github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/diskcache"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/afero"
)

func TestWebDAV(t *testing.T) {
	root := t.TempDir()
	perm := users.Permissions{Create: true, Modify: true, Rename: true, Delete: true, Download: true}
	key := []byte("dav-test-key")
	st := scopedUserStorage(t, root, perm, key)
	server := &settings.Server{Root: root, EnableWebDAV: true, BaseURL: "/files"}
	if err := st.Settings.Save(&settings.Settings{Key: key, AuthMethod: fbAuth.MethodJSONAuth,
		Rules: []rules.Rule{{Path: "/secret", Allow: false}, {Path: "/tree/blocked", Allow: false}}}); err != nil {
		t.Fatal(err)
	}
	if err := st.Auth.Save(&fbAuth.JSONAuth{}); err != nil {
		t.Fatal(err)
	}
	u, err := st.Users.Get(root, false, "u")
	if err != nil {
		t.Fatal(err)
	}
	u.Password, err = users.HashPwd("dav-password")
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Users.Save(u); err != nil {
		t.Fatal(err)
	}
	st.Users = &customFSUser{Store: st.Users, fs: files.NewFs(afero.NewOsFs(), root, false)}
	// Test the real router, including BaseURL handling and method dispatch.
	cache := diskcache.New(afero.NewOsFs(), t.TempDir())
	h, err := NewHandler(nil, cache, nil, st, server, os.DirFS(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	request := func(method, name, body, destination string, auth bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "/files/dav"+name, strings.NewReader(body))
		if auth {
			r.SetBasicAuth("u", "dav-password")
		}
		if destination != "" {
			r.Header.Set("Destination", "/files/dav"+destination)
		}
		r.Header.Set("Depth", "1")
		if method == "COPY" || method == "MOVE" {
			r.Header.Del("Depth")
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	assert := func(w *httptest.ResponseRecorder, status int) {
		t.Helper()
		if w.Code != status {
			t.Fatalf("got %d, want %d: %s", w.Code, status, w.Body.String())
		}
	}
	assert(request("PROPFIND", "/", "", "", false), 401)
	assert(request("MKCOL", "/folder", "", "", true), 201)
	assert(request("PUT", "/folder/你好.txt", "hello", "", true), 201)
	w := request("GET", "/folder/你好.txt", "", "", true)
	assert(w, 200)
	if w.Body.String() != "hello" {
		t.Fatal(w.Body.String())
	}
	assert(request("COPY", "/folder/你好.txt", "", "/copy.txt", true), 201)
	assert(request("MOVE", "/copy.txt", "", "/moved.txt", true), 201)
	assert(request("DELETE", "/moved.txt", "", "", true), 204)
	assert(request("COPY", "/folder/你好.txt", "", "/folder", true), 403)
	info, err := u.Fs.Stat("/folder/你好.txt")
	if err != nil {
		t.Fatal(err)
	}
	previewKey := previewCacheKey(&files.FileInfo{Fs: u.Fs, Path: "/folder/你好.txt", ModTime: info.ModTime()}, PreviewSizeThumb)
	if err := cache.Store(context.Background(), previewKey, []byte("old thumbnail")); err != nil {
		t.Fatal(err)
	}
	// Locks survive individual requests; a write requires the returned token.
	lockRequest := func(method, body, token string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "/files/dav/folder/你好.txt", strings.NewReader(body))
		r.Header.Set("X-Auth", signToken(t, perm, key, st))
		r.Header.Set("Depth", "0")
		if method == "UNLOCK" {
			r.Header.Set("Lock-Token", token)
		} else if token != "" {
			r.Header.Set("If", "("+token+")")
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	lock := lockRequest("LOCK", `<D:lockinfo xmlns:D="DAV:"><D:lockscope><D:exclusive/></D:lockscope><D:locktype><D:write/></D:locktype></D:lockinfo>`, "")
	assert(lock, 200)
	token := lock.Header().Get("Lock-Token")
	if token == "" {
		t.Fatal("missing lock token")
	}
	assert(lockRequest("PUT", "locked", ""), 423)
	assert(lockRequest("PUT", "updated", token), 201)
	if _, exists, err := cache.Load(context.Background(), previewKey); err != nil || exists {
		t.Fatalf("stale preview remains: %v", err)
	}
	assert(lockRequest("LOCK", "", token), 200)
	assert(lockRequest("UNLOCK", "", token), 204)
	if err := os.WriteFile(filepath.Join(root, "secret"), []byte("hidden"), 0600); err != nil {
		t.Fatal(err)
	}
	w = request("PROPFIND", "/", "", "", true)
	assert(w, 207)
	if strings.Contains(w.Body.String(), "secret") || !strings.Contains(w.Body.String(), "/files/dav/folder") {
		t.Fatal(w.Body.String())
	}
	assert(request("GET", "/secret", "", "", true), 403)
	assert(request("DELETE", "/", "", "", true), 403)
	if err := os.MkdirAll(filepath.Join(root, "tree"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "tree", "blocked"), []byte("protected"), 0600); err != nil {
		t.Fatal(err)
	}
	assert(request("DELETE", "/tree", "", "", true), 403)
	assert(request("MOVE", "/tree", "", "/new-tree", true), 403)
	assert(request("COPY", "/tree", "", "/new-tree", true), 403)
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link")); err == nil {
		if w := request("GET", "/link", "", "", true); w.Code < 400 {
			t.Fatal("symlink escaped scope")
		}
	}
	// Updating permissions takes effect on the next authenticated request.
	u.Perm = users.Permissions{}
	if err := st.Users.Save(u); err != nil {
		t.Fatal(err)
	}
	for _, method := range []string{"GET", "HEAD", "PUT", "MKCOL", "DELETE", "COPY", "MOVE", "LOCK", "PROPPATCH"} {
		assert(request(method, "/folder", "", "/other", true), http.StatusForbidden)
	}
}

func TestWebDAVRejectsBasicAuthBypass(t *testing.T) {
	key := []byte("dav-test-key")
	st := scopedUserStorage(t, t.TempDir(), users.Permissions{}, key)
	if err := st.Auth.Save(&fbAuth.JSONAuth{ReCaptcha: &fbAuth.ReCaptcha{Secret: "required"}}); err != nil {
		t.Fatal(err)
	}
	h := handle(newWebDAVHandler(nil), "", st, &settings.Server{EnableWebDAV: true})
	for _, method := range []settings.AuthMethod{fbAuth.MethodJSONAuth, fbAuth.MethodProxyAuth, fbAuth.MethodNoAuth} {
		if err := st.Settings.Save(&settings.Settings{Key: key, AuthMethod: method}); err != nil {
			t.Fatal(err)
		}
		r := httptest.NewRequest("PROPFIND", "/dav/", nil)
		r.SetBasicAuth("u", "pw")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 401 {
			t.Fatalf("auth method %s: %d", method, w.Code)
		}
	}
}
