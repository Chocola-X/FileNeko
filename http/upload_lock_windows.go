package fbhttp

import "golang.org/x/sys/windows"

func lockUploadFD(fd uintptr) (func(), error) {
	var overlapped windows.Overlapped
	if err := windows.LockFileEx(windows.Handle(fd), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &overlapped); err != nil {
		return nil, err
	}
	return func() { _ = windows.UnlockFileEx(windows.Handle(fd), 0, 1, 0, &overlapped) }, nil
}
