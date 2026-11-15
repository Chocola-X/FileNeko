package fbhttp

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/afero"
)

func davSecurityFixture(t *testing.T, root string) (http.Handler, string, *storage.Storage) {
	t.Helper()
	key := []byte("dav-content-security")
	perm := users.Permissions{Download: true, Create: true, Modify: true, Delete: true, Rename: true}
	st := scopedUserStorage(t, root, perm, key)
	st.Users = &customFSUser{Store: st.Users, fs: files.NewFs(afero.NewOsFs(), root, false)}
	h, err := NewHandler(nil, nil, nil, st, &settings.Server{Root: root, EnableWebDAV: true}, os.DirFS(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	return h, signToken(t, perm, key, st), st
}

func davSecurityRequest(h http.Handler, token, method, name, body string, headers map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "/dav"+name, strings.NewReader(body))
	r.Header.Set("X-Auth", token)
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestWebDAVFileResponsesCannotExecuteUploadedContent(t *testing.T) {
	root := t.TempDir()
	payload := `<script src="/dav/payload.js"></script>`
	for _, name := range []string{"payload.html", "payload.svg", "payload.js"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(payload), 0600); err != nil {
			t.Fatal(err)
		}
	}
	h, token, _ := davSecurityFixture(t, root)
	for _, name := range []string{"payload.html", "payload.svg", "payload.js"} {
		for _, method := range []string{"GET", "HEAD"} {
			w := davSecurityRequest(h, token, method, "/"+name, "", nil)
			if w.Code != 200 {
				t.Fatalf("%s %s: %d", method, name, w.Code)
			}
			csp := w.Header().Get("Content-Security-Policy")
			if !strings.Contains(csp, "script-src 'none'") || !strings.Contains(csp, "sandbox") || w.Header().Get("X-Content-Type-Options") != "nosniff" {
				t.Fatalf("uploaded content can execute: CSP=%q nosniff=%q", csp, w.Header().Get("X-Content-Type-Options"))
			}
			if !strings.HasPrefix(w.Header().Get("Content-Disposition"), "attachment") {
				t.Fatal("uploaded content is served inline")
			}
			if method == "GET" && w.Body.String() != payload {
				t.Fatal("security headers changed download content")
			}
		}
	}
}
