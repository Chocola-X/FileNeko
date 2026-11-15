package files

import (
	"context"
	"io"
	"io/fs"
	"path"

	"github.com/spf13/afero"
)

// DirectoryStats counts accessible descendants, excluding the root and symbolic
// links. Size is the sum of file lengths, not allocated disk blocks; hard links
// are counted once per directory entry. Nothing is persisted.
type DirectoryStats struct {
	Size       int64 `json:"size"`
	NumFiles   int64 `json:"numFiles"`
	NumDirs    int64 `json:"numDirs"`
	Incomplete bool  `json:"incomplete"`
}

// ScanDirectory reads only metadata, in bounded, unsorted batches. Each opened
// directory goes through the user's filesystem guard. Readdir supplies child
// metadata without opening files or dereferencing their symbolic links.
// Cancellation is checked before filesystem operations and between entries.
func ScanDirectory(ctx context.Context, afs afero.Fs, root string, allowed func(string) bool,
	progress func(DirectoryStats) error,
) (DirectoryStats, error) {
	var stats DirectoryStats
	if err := ctx.Err(); err != nil {
		return stats, err
	}
	if !allowed(root) {
		return stats, fs.ErrPermission
	}
	info, err := lstatIfPossible(afs, root)
	if err != nil {
		return stats, err
	}
	if !info.IsDir() || IsSymlink(info.Mode()) {
		return stats, fs.ErrInvalid
	}

	var walk func(string) error
	walk = func(dirname string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		dir, err := afs.Open(dirname)
		if err != nil {
			stats.Incomplete = true
			return nil
		}
		defer dir.Close()
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			entries, readErr := dir.Readdir(256)
			for _, entry := range entries {
				if err := ctx.Err(); err != nil {
					return err
				}
				if IsSymlink(entry.Mode()) {
					continue
				}
				name := path.Join(dirname, entry.Name())
				if !allowed(name) {
					continue
				}
				if entry.IsDir() {
					stats.NumDirs++
				} else {
					stats.NumFiles++
					stats.Size += entry.Size()
				}
				if progress != nil {
					if err := progress(stats); err != nil {
						return err
					}
				}
				if entry.IsDir() {
					if err := walk(name); err != nil {
						return err
					}
				}
			}
			if readErr != nil && readErr != io.EOF {
				stats.Incomplete = true
			}
			if readErr != nil || len(entries) == 0 {
				return nil
			}
		}
	}
	err = walk(root)
	return stats, err
}
