package files

import (
	"errors"
	"os"
	"path/filepath"
)

// ResolvePath resolves links for existing files and for a future write target.
// Missing suffixes are preserved; dangling symlinks are followed rather than
// mistaken for new regular files. This is a path check, not an OS sandbox.
func ResolvePath(name string) (string, error) {
	target := filepath.Clean(name)
	suffix := ""
	for hops := 0; hops < maxSymlinkHops; hops++ {
		resolved, err := filepath.EvalSymlinks(target)
		if err == nil {
			return filepath.Join(resolved, suffix), nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		info, statErr := os.Lstat(target)
		if statErr == nil && info.Mode()&os.ModeSymlink != 0 {
			dest, err := os.Readlink(target)
			if err != nil {
				return "", err
			}
			if !filepath.IsAbs(dest) {
				parent, err := filepath.EvalSymlinks(filepath.Dir(target))
				if err != nil {
					return "", err
				}
				dest = filepath.Join(parent, dest)
			}
			target = filepath.Clean(dest)
			continue
		}
		parent := filepath.Dir(target)
		if parent == target {
			return "", err
		}
		suffix = filepath.Join(filepath.Base(target), suffix)
		target = parent
	}
	return "", os.ErrPermission
}
