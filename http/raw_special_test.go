package fbhttp

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/afero"
)

type specialInfo struct {
	os.FileInfo
	mode os.FileMode
}

func (i specialInfo) Mode() os.FileMode { return i.mode }

type specialFs struct {
	afero.Fs
	mode os.FileMode
}

func (f specialFs) Stat(name string) (os.FileInfo, error) {
	info, err := f.Fs.Stat(name)
	if err == nil && strings.HasSuffix(name, "special") {
		return specialInfo{info, f.mode}, nil
	}
	return info, err
}
func (f specialFs) Open(name string) (afero.File, error) {
	if strings.HasSuffix(name, "special") {
		panic("must not open a FIFO, device or socket")
	}
	return f.Fs.Open(name)
}

// GHSA-8q5j-8wcr-8v2v: never reach the potentially blocking Open, on either
// the archive path or the shared single-file download path.
func TestDownloadsSkipSpecialFiles(t *testing.T) {
	for _, mode := range []os.FileMode{os.ModeNamedPipe, os.ModeDevice, os.ModeSocket} {
		base := afero.NewMemMapFs()
		if err := afero.WriteFile(base, "/special", nil, 0644); err != nil {
			t.Fatal(err)
		}
		fs := specialFs{base, mode}
		d := &data{user: &users.User{Fs: fs}, settings: &settings.Settings{}, server: &settings.Server{}}
		entries, err := getFiles(d, "/", "/")
		if err != nil || len(entries) != 0 {
			t.Fatalf("special file included: %v %v", entries, err)
		}
		status, err := rawFileHandler(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), &files.FileInfo{Fs: fs, Path: "/special"})
		if err != nil || status != 400 {
			t.Fatalf("special download: %d %v", status, err)
		}
	}
}
