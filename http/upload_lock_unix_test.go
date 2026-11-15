//go:build linux || darwin || dragonfly || freebsd || netbsd || openbsd

package fbhttp

import (
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
	"testing"
)

func TestUploadFileLockCoordinatesIndependentHandles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "upload")
	a, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := os.OpenFile(path, os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	unlock, err := lockUploadFD(a.Fd())
	if err != nil {
		t.Fatal(err)
	}
	if release, err := lockUploadFD(b.Fd()); err == nil {
		release()
		unlock()
		t.Fatal("second handle acquired an active upload lock")
	}
	unlock()
	release, err := lockUploadFD(b.Fd())
	if err != nil {
		t.Fatal(err)
	}
	release()
}

func TestUploadLockUnwrapsScopedFile(t *testing.T) {
	root := t.TempDir()
	fs := files.NewScopedFs(afero.NewBasePathFs(afero.NewOsFs(), root), "/")
	file, err := fs.OpenFile("/upload", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	fd, ok := uploadFileDescriptor(file)
	if !ok {
		t.Fatal("scoped OS file descriptor was hidden")
	}
	unlock, err := lockUploadFD(fd)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	other, err := os.OpenFile(filepath.Join(root, "upload"), os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	if release, err := lockUploadFD(other.Fd()); err == nil {
		release()
		t.Fatal("scope wrappers bypassed kernel lock")
	}
}
