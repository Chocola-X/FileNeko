package fbhttp

import (
	"fmt"
	"hash/fnv"
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
	token := signToken(t, perm, key, st)

	// Choose two different paths that would share a lock in the old handler.
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	var buckets [256]string
	var firstPath, otherPath string
	for i := 0; i <= len(buckets); i++ {
		name := fmt.Sprintf("upload-%d.bin", i)
		hash := fnv.New32a()
		_, _ = hash.Write([]byte(filepath.Join(resolvedRoot, name)))
		bucket := hash.Sum32() % uint32(len(buckets))
		if buckets[bucket] != "" {
			firstPath, otherPath = "/"+buckets[bucket], "/"+name
			break
		}
		buckets[bucket] = name
	}
	for _, path := range []string{firstPath, otherPath} {
		post := httptest.NewRequest("POST", path, nil)
		post.Header.Set("X-Auth", token)
		post.Header.Set("Upload-Length", "5")
		rec := httptest.NewRecorder()
		handle(tusPostHandler(cache), "", st, &settings.Server{}).ServeHTTP(rec, post)
		if rec.Code != 201 {
			t.Fatalf("create %s: %d %s", path, rec.Code, rec.Body)
		}
	}
	patch := handle(tusPatchHandler(cache), "", st, &settings.Server{})
	request := func(path string, body io.Reader) *httptest.ResponseRecorder {
		req := httptest.NewRequest("PATCH", path, body)
		req.Header.Set("X-Auth", token)
		req.Header.Set("Upload-Offset", "0")
		req.Header.Set("Content-Type", "application/offset+octet-stream")
		rec := httptest.NewRecorder()
		patch.ServeHTTP(rec, req)
		return rec
	}
	body := &pausedUpload{started: make(chan struct{}), release: make(chan struct{}), reader: strings.NewReader("hello")}
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() { done <- request(firstPath, body) }()
	select {
	case <-body.started:
	case <-time.After(5 * time.Second):
		close(body.release)
		t.Fatal("PATCH never reached body")
	}
	second := request(firstPath, strings.NewReader("other"))
	independent := request(otherPath, strings.NewReader("world"))
	close(body.release)
	first := <-done
	if first.Code != 204 || second.Code != 409 || independent.Code != 204 {
		t.Fatalf("PATCH statuses: first=%d same file=%d different file=%d", first.Code, second.Code, independent.Code)
	}
	for path, expected := range map[string]string{firstPath: "hello", otherPath: "world"} {
		content, err := os.ReadFile(filepath.Join(root, filepath.Base(path)))
		if err != nil || string(content) != expected {
			t.Fatalf("stored upload %s: %q %v", path, content, err)
		}
	}
}
