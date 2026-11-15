package files

import (
	"errors"
	"os"
	"testing"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/spf13/afero"
)

type staleContentInfo struct{ os.FileInfo }

func (staleContentInfo) Size() int64 { return 1 }

type staleContentFs struct{ afero.Fs }

func (f staleContentFs) Stat(name string) (os.FileInfo, error) {
	info, err := f.Fs.Stat(name)
	if err != nil {
		return nil, err
	}
	return staleContentInfo{info}, nil
}

// GHSA-7xqm-7738-642x / #5294: the read must stay bounded even when file
// metadata was collected before an attacker grew the file.
func TestTextContentReadDoesNotTrustStaleSize(t *testing.T) {
	fs := afero.NewMemMapFs()
	f, err := fs.Create("/growing.txt")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(MaxTextSize + 1); err != nil {
		t.Fatal(err)
	}
	f.Close()
	_, err = NewFileInfo(&FileOptions{Fs: staleContentFs{fs}, Path: "/growing.txt", Content: true, Expand: true, Modify: true, Checker: allowAllChecker{}})
	if !errors.Is(err, fberrors.ErrFileTooLarge) {
		t.Fatalf("unbounded text read: %v", err)
	}
}
