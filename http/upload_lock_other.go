//go:build !linux && !darwin && !dragonfly && !freebsd && !netbsd && !openbsd && !windows

package fbhttp

import "errors"

func lockUploadFD(_ uintptr) (func(), error) {
	return nil, errors.New("upload file locking is unsupported on this platform")
}
