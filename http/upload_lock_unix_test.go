//go:build linux || darwin || dragonfly || freebsd || netbsd || openbsd

package fbhttp

import (
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
