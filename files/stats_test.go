package files

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestScanDirectoryTotalsAndRules(t *testing.T) {
	afs := afero.NewMemMapFs()
	for name, content := range map[string]string{
		"/root/one.txt":           "123",
		"/root/sub/two.txt":       "12345",
		"/root/.hidden":           "12",
		"/root/denied/secret.txt": "not accessible",
	} {
		if err := afs.MkdirAll(filepath.Dir(name), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := afero.WriteFile(afs, name, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := afs.Mkdir("/root/empty", 0o755); err != nil {
		t.Fatal(err)
	}
	var updates []DirectoryStats
	stats, err := ScanDirectory(context.Background(), afs, "/root", func(name string) bool {
		return !strings.HasPrefix(name, "/root/denied")
	}, func(stats DirectoryStats) error {
		updates = append(updates, stats)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if stats != (DirectoryStats{Size: 10, NumFiles: 3, NumDirs: 2}) {
		t.Fatalf("unexpected totals: %+v", stats)
	}
	for i := 1; i < len(updates); i++ {
		if updates[i].Size < updates[i-1].Size || updates[i].NumFiles < updates[i-1].NumFiles {
			t.Fatal("progress must be cumulative")
		}
	}
}

func TestScanDirectoryDoesNotFollowLinksOrReadContents(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "sub", "file"), []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	for name, target := range map[string]string{
		"outside": t.TempDir(), "loop": root, "duplicate": filepath.Join(root, "sub"),
	} {
		if err := os.Symlink(target, filepath.Join(root, name)); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
	}
	afs := &statsMetadataOnlyFs{Fs: NewScopedFs(afero.NewOsFs(), root)}
	stats, err := ScanDirectory(context.Background(), afs, "/", func(string) bool { return true }, nil)
	if err != nil || stats != (DirectoryStats{Size: 4, NumFiles: 1, NumDirs: 1}) {
		t.Fatalf("stats = %+v, err = %v", stats, err)
	}
	if afs.openFiles != 0 {
		t.Fatal("scan opened file contents")
	}
	if afs.openDirs != afs.closedDirs {
		t.Fatalf("unclosed directory handles: %d/%d", afs.closedDirs, afs.openDirs)
	}
}

type statsMetadataOnlyFs struct {
	afero.Fs
	openFiles, openDirs, closedDirs int
}

func (s *statsMetadataOnlyFs) Open(name string) (afero.File, error) {
	f, err := s.Fs.Open(name)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if !info.IsDir() {
		s.openFiles++
	}
	s.openDirs++
	return &statsMetadataOnlyFile{File: f, closed: &s.closedDirs}, nil
}

type statsMetadataOnlyFile struct {
	afero.File
	closed *int
}

func (f *statsMetadataOnlyFile) Read([]byte) (int, error) {
	panic("a directory scan must not read file contents")
}

func (f *statsMetadataOnlyFile) Readdir(n int) ([]os.FileInfo, error) {
	if n <= 0 || n > 256 {
		panic("a directory scan must use bounded batches")
	}
	return f.File.Readdir(n)
}

func (f *statsMetadataOnlyFile) Close() error {
	*f.closed++
	return f.File.Close()
}

func TestScanDirectoryCancellationClosesHandlesAndStopsVisiting(t *testing.T) {
	mem := afero.NewMemMapFs()
	if err := mem.MkdirAll("/root/sub", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := afero.WriteFile(mem, "/root/sub/file", []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	afs := &statsMetadataOnlyFs{Fs: mem}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	visits := 0
	_, err := ScanDirectory(ctx, afs, "/root", func(string) bool { return true }, func(DirectoryStats) error {
		visits++
		cancel()
		return nil
	})
	if !errors.Is(err, context.Canceled) || visits != 1 || afs.openDirs != 1 || afs.closedDirs != 1 {
		t.Fatalf("err=%v visits=%d open=%d closed=%d", err, visits, afs.openDirs, afs.closedDirs)
	}
	_, err = ScanDirectory(ctx, afs, "/root", func(string) bool { return true }, nil)
	if !errors.Is(err, context.Canceled) || afs.openDirs != 1 {
		t.Fatalf("canceled scan accessed the filesystem: %v", err)
	}
}

type statsUnreadableFs struct{ afero.Fs }

func (s statsUnreadableFs) Open(name string) (afero.File, error) {
	if strings.HasSuffix(name, "/blocked") {
		return nil, fs.ErrPermission
	}
	return s.Fs.Open(name)
}

func TestScanDirectoryReportsUnreadableDescendants(t *testing.T) {
	afs := afero.NewMemMapFs()
	if err := afs.MkdirAll("/root/blocked", 0o755); err != nil {
		t.Fatal(err)
	}
	stats, err := ScanDirectory(context.Background(), statsUnreadableFs{afs}, "/root", func(string) bool { return true }, nil)
	if err != nil || !stats.Incomplete || stats.NumDirs != 1 {
		t.Fatalf("stats=%+v err=%v", stats, err)
	}
}
