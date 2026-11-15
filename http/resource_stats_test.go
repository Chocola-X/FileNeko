package fbhttp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/afero"
)

func statsTestHandler(t *testing.T, root string) (http.Handler, string, *storage.Storage) {
	t.Helper()
	key := []byte("stats-test-signing-key")
	perm := users.Permissions{} // metadata does not require file downloads
	st := scopedUserStorage(t, root, perm, key)
	token := signToken(t, perm, key, st)
	return handle(newDirectoryStatsHandler(), "/api/directory-stats", st, &settings.Server{}), token, st
}

func TestDirectoryStatsStreamTotalsAndAccess(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{"file": "123", ".hidden": "12", "sub/nested": "1234", "secret": "denied"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	handler, token, st := statsTestHandler(t, root)
	s, err := st.Settings.Get()
	if err != nil {
		t.Fatal(err)
	}
	s.Rules = []rules.Rule{{Path: "/secret", Allow: false}}
	if err := st.Settings.Save(s); err != nil {
		t.Fatal(err)
	}
	u, err := st.Users.Get("", false, uint(1))
	if err != nil {
		t.Fatal(err)
	}
	u.HideDotfiles = true
	if err := st.Users.Save(u); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/directory-stats/", nil)
	req.Header.Set("X-Auth", token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Header().Get("Cache-Control") != "no-store" || rec.Header().Get("X-Accel-Buffering") != "no" || !rec.Flushed {
		t.Fatalf("status=%d headers=%v flushed=%v body=%s", rec.Code, rec.Header(), rec.Flushed, rec.Body.String())
	}
	decoder := json.NewDecoder(rec.Body)
	var initial, final directoryStatsUpdate
	if err := decoder.Decode(&initial); err != nil || initial.Done || initial.Size != 0 {
		t.Fatalf("initial=%+v err=%v", initial, err)
	}
	for {
		var update directoryStatsUpdate
		if err := decoder.Decode(&update); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		final = update
	}
	if !final.Done || final.Incomplete || final.Size != 9 || final.NumFiles != 3 || final.NumDirs != 1 {
		t.Fatalf("unexpected totals: %+v", final)
	}
	for path, want := range map[string]int{"/": http.StatusUnauthorized, "/secret": http.StatusForbidden, "/file": http.StatusBadRequest, "/missing": http.StatusNotFound} {
		req := httptest.NewRequest(http.MethodGet, "/api/directory-stats"+path, nil)
		if path != "/" {
			req.Header.Set("X-Auth", token)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("%s: status %d, want %d; %s", path, rec.Code, want, rec.Body.String())
		}
	}
}

func TestDirectoryStatsRejectsSymbolicLinkRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "real"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "real"), filepath.Join(root, "link")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	handler, token, _ := statsTestHandler(t, root)
	req := httptest.NewRequest(http.MethodGet, "/api/directory-stats/link", nil)
	req.Header.Set("X-Auth", token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

type statsBlockingFs struct {
	afero.Fs
	entered chan struct{}
	release chan struct{}
}

func (s *statsBlockingFs) Open(name string) (afero.File, error) {
	f, err := s.Fs.Open(name)
	if err != nil {
		return nil, err
	}
	return &statsBlockingFile{File: f, fs: s}, nil
}

type statsBlockingFile struct {
	afero.File
	fs   *statsBlockingFs
	once sync.Once
}

func (f *statsBlockingFile) Readdir(n int) ([]os.FileInfo, error) {
	f.once.Do(func() {
		f.fs.entered <- struct{}{}
		<-f.fs.release
	})
	return f.File.Readdir(n)
}

func statsWait(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for scan")
	}
}

func TestDirectoryStatsStopsAfterHTTPDisconnect(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "must-not-visit"), 0o755); err != nil {
		t.Fatal(err)
	}
	handler, token, st := statsTestHandler(t, root)
	blocking := &statsBlockingFs{Fs: afero.NewOsFs(), entered: make(chan struct{}, 1), release: make(chan struct{})}
	st.Users.(*customFSUser).fs = afero.NewBasePathFs(blocking, root)
	finished := make(chan struct{})
	contexts := make(chan context.Context, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contexts <- r.Context()
		defer close(finished)
		handler.ServeHTTP(w, r)
	}))
	defer server.Close()
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(blocking.release) }) }
	defer release()
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/api/directory-stats/", http.NoBody)
	req.Header.Set("X-Auth", token)
	response, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	statsWait(t, blocking.entered)
	ctx := <-contexts
	if err := response.Body.Close(); err != nil {
		t.Fatal(err)
	}
	statsWait(t, ctx.Done())
	// A running filesystem call cannot be preempted; once it returns, the
	// canceled request must finish without traversing another directory.
	release()
	statsWait(t, finished)
	select {
	case <-blocking.entered:
		t.Fatal("disconnected scan continued into a child directory")
	default:
	}
}

type statsNotifyWriter struct {
	*httptest.ResponseRecorder
	flushed chan struct{}
	once    sync.Once
}

func (w *statsNotifyWriter) Flush() {
	w.ResponseRecorder.Flush()
	w.once.Do(func() { close(w.flushed) })
}

func TestDirectoryStatsLimitsConcurrencyAndCancelsQueuedRequest(t *testing.T) {
	root := t.TempDir()
	handler, token, st := statsTestHandler(t, root)
	blocking := &statsBlockingFs{Fs: afero.NewOsFs(), entered: make(chan struct{}, 3), release: make(chan struct{})}
	st.Users.(*customFSUser).fs = afero.NewBasePathFs(blocking, root)
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(blocking.release) }) }
	defer release()
	run := func(ctx context.Context) (*statsNotifyWriter, chan struct{}) {
		req := httptest.NewRequest(http.MethodGet, "/api/directory-stats/", nil).WithContext(ctx)
		req.Header.Set("X-Auth", token)
		writer := &statsNotifyWriter{ResponseRecorder: httptest.NewRecorder(), flushed: make(chan struct{})}
		finished := make(chan struct{})
		go func() { defer close(finished); handler.ServeHTTP(writer, req) }()
		return writer, finished
	}
	_, first := run(context.Background())
	statsWait(t, blocking.entered)
	_, second := run(context.Background())
	statsWait(t, blocking.entered)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	queued, third := run(ctx)
	statsWait(t, queued.flushed)
	cancel()
	statsWait(t, third)
	var update directoryStatsUpdate
	if err := json.NewDecoder(queued.Body).Decode(&update); err != nil || !update.Queued || update.Done {
		t.Fatalf("update=%+v err=%v", update, err)
	}
	select {
	case <-blocking.entered:
		t.Fatal("queued/canceled request started a third scan")
	default:
	}
	release()
	statsWait(t, first)
	statsWait(t, second)
}
