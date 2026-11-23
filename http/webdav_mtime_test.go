package fbhttp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

func TestWebDAVMtime(t *testing.T) {
	root := t.TempDir()
	perm := users.Permissions{Create: true, Modify: true, Download: true}
	key := []byte("dav-mtime-test")
	st := scopedUserStorage(t, root, perm, key)
	token := signToken(t, perm, key, st)
	handler, err := NewHandler(nil, nil, nil, st, &settings.Server{
		Root: root, EnableWebDAV: true, BaseURL: "/files",
	}, os.DirFS(root))
	if err != nil {
		t.Fatal(err)
	}

	request := func(method, path, body, mtime string, status int) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, "/files/dav/"+path, strings.NewReader(body))
		r.Header.Set("X-Auth", token)
		r.Header.Set("Depth", "0")
		if mtime != "" {
			r.Header.Set(davMtimeHeader, mtime)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != status {
			t.Fatalf("%s %s: got %d, want %d: %s", method, path, w.Code, status, w.Body)
		}
		return w
	}
	checkMtime := func(path string, seconds int64) {
		t.Helper()
		info, err := os.Stat(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		if want := time.Unix(seconds, 0); !info.ModTime().Equal(want) {
			t.Fatalf("%s: mtime = %v, want %v", path, info.ModTime(), want)
		}
	}
	checkAccepted := func(w *httptest.ResponseRecorder, want string) {
		t.Helper()
		if got := w.Header().Get(davMtimeHeader); got != want {
			t.Fatalf("X-OC-Mtime = %q, want %q", got, want)
		}
	}

	w := request("MKCOL", "folder/", "", "1600000000", http.StatusCreated)
	checkAccepted(w, "accepted")
	checkMtime("folder", 1600000000)
	w = request(http.MethodPut, "folder/file.txt", "data", "1700000000", http.StatusCreated)
	checkAccepted(w, "accepted")
	checkMtime("folder/file.txt", 1700000000)
	head := request(http.MethodHead, "folder/file.txt", "", "", http.StatusOK)
	if w.Header().Get("ETag") == "" || w.Header().Get("ETag") != head.Header().Get("ETag") {
		t.Fatal("PUT and HEAD ETags differ")
	}
	w = request(http.MethodPut, "folder/file.txt", "", "1500000000", http.StatusCreated)
	checkAccepted(w, "accepted")
	checkMtime("folder/file.txt", 1500000000)
	w = request(http.MethodPut, "folder/file.txt", "replace", "bad", http.StatusBadRequest)
	checkAccepted(w, "")
	checkMtime("folder/file.txt", 1500000000)
	if info, err := os.Stat(filepath.Join(root, "folder", "file.txt")); err != nil || info.Size() != 0 {
		t.Fatalf("invalid header changed the empty file: %v, %v", info, err)
	}
	checkAccepted(request(http.MethodPut, "folder/plain.txt", "plain", "", http.StatusCreated), "")

	broken := httptest.NewRequest(http.MethodPut, "/files/dav/folder/broken.txt", iotest.ErrReader(errors.New("upload interrupted")))
	broken.Header.Set("X-Auth", token)
	broken.Header.Set(davMtimeHeader, "1500000000")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, broken)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("interrupted PUT returned %d", w.Code)
	}
	checkAccepted(w, "")

	patch := `<D:propertyupdate xmlns:D="DAV:"><D:set><D:prop><D:lastmodified>1600000000</D:lastmodified></D:prop></D:set></D:propertyupdate>`
	w = request("PROPPATCH", "folder/", patch, "", http.StatusMultiStatus)
	if !strings.Contains(w.Body.String(), "HTTP/1.1 200 OK") {
		t.Fatalf("directory timestamp update failed: %s", w.Body)
	}
	checkMtime("folder", 1600000000)
	w = request("PROPFIND", "folder/", "", "", http.StatusMultiStatus)
	if !strings.Contains(w.Body.String(), time.Unix(1600000000, 0).UTC().Format(http.TimeFormat)) {
		t.Fatalf("PROPFIND omitted the stored directory time: %s", w.Body)
	}
	mixed := strings.Replace(patch, "1600000000</D:lastmodified>", "1700000000</D:lastmodified><D:unknown/>", 1)
	w = request("PROPPATCH", "folder/", mixed, "", http.StatusMultiStatus)
	if !strings.Contains(w.Body.String(), "424 Failed Dependency") {
		t.Fatalf("mixed patch did not fail atomically: %s", w.Body)
	}
	checkMtime("folder", 1600000000)

	request("LOCK", "folder/", `<D:lockinfo xmlns:D="DAV:"><D:lockscope><D:exclusive/></D:lockscope><D:locktype><D:write/></D:locktype></D:lockinfo>`, "", http.StatusOK)
	request("PROPPATCH", "folder/", patch, "", 423)
	checkMtime("folder", 1600000000)

	user, err := st.Users.Get(root, false, "u")
	if err != nil {
		t.Fatal(err)
	}
	user.Perm.Modify = false
	if err := st.Users.Save(user); err != nil {
		t.Fatal(err)
	}
	request("PROPPATCH", "folder/plain.txt", patch, "", http.StatusForbidden)
}
