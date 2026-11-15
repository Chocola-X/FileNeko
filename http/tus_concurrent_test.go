package fbhttp

import (
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

type pausedUpload struct {
	started, release chan struct{}
	reader           io.Reader
	first            bool
}

func (b *pausedUpload) Read(p []byte) (int, error) {
	if !b.first {
		b.first = true
		close(b.started)
		<-b.release
	}
	return b.reader.Read(p)
}
func (b *pausedUpload) Close() error { return nil }

// GHSA-4r8p-gqj2-mwgm: hold the first PATCH inside its body read and submit
// a second PATCH at the same offset. Exactly one can write or complete.
func TestTusConcurrentPatchCannotExceedLength(t *testing.T) {
	root := t.TempDir()
	key := []byte("key")
	perm := users.Permissions{Create: true, Modify: true}
	st := scopedUserStorage(t, root, perm, key)
	cache := newMemoryUploadCache()
	defer cache.Close()
	token := signToken(t, perm, key)
	post := httptest.NewRequest("POST", "/race.bin", nil)
	post.Header.Set("X-Auth", token)
	post.Header.Set("Upload-Length", "5")
	rec := httptest.NewRecorder()
	handle(tusPostHandler(cache), "", st, &settings.Server{}).ServeHTTP(rec, post)
	if rec.Code != 201 {
		t.Fatalf("create: %d %s", rec.Code, rec.Body)
	}
	patch := handle(tusPatchHandler(cache), "", st, &settings.Server{})
	request := func(body io.Reader) *httptest.ResponseRecorder {
		req := httptest.NewRequest("PATCH", "/race.bin", body)
		req.Header.Set("X-Auth", token)
		req.Header.Set("Upload-Offset", "0")
		req.Header.Set("Content-Type", "application/offset+octet-stream")
		rec := httptest.NewRecorder()
		patch.ServeHTTP(rec, req)
		return rec
	}
	body := &pausedUpload{started: make(chan struct{}), release: make(chan struct{}), reader: strings.NewReader("hello")}
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() { done <- request(body) }()
	select {
	case <-body.started:
	case <-time.After(5 * time.Second):
		close(body.release)
		t.Fatal("PATCH never reached body")
	}
	second := request(strings.NewReader("other"))
	close(body.release)
	first := <-done
	if first.Code != 204 || second.Code != 409 {
		t.Fatalf("PATCH statuses: %d, %d", first.Code, second.Code)
	}
	content, err := os.ReadFile(filepath.Join(root, "race.bin"))
	if err != nil || string(content) != "hello" {
		t.Fatalf("stored upload: %q %v", content, err)
	}
}
