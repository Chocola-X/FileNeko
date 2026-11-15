//go:build linux || darwin || dragonfly || freebsd || netbsd || openbsd

package fbhttp

import "golang.org/x/sys/unix"

func lockUploadFD(fd uintptr) (func(), error) {
	if err := unix.Flock(int(fd), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		return nil, err
	}
	return func() { _ = unix.Flock(int(fd), unix.LOCK_UN) }, nil
}
