package fbhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	fbAuth "github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
)

// Locks are shared across requests and users: two accounts can share a scope.
// Use physical paths for lock keys so different scopes cannot alias a lock.
func newWebDAVHandler(cache FileCache) handleFunc {
	locks := newDAVLockRegistry()
	serve := func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		prefix := d.server.BaseURL + "/dav"
		r = r.Clone(r.Context())
		r.URL.Path = d.server.BaseURL + r.URL.Path
		r.URL.RawPath = ""
		src := slashClean(strings.TrimPrefix(r.URL.Path, prefix))
		if !d.Check(src) {
			return http.StatusForbidden, nil
		}
		p := d.user.Perm
		switch r.Method {
		case "GET", "HEAD":
			if !p.Download {
				return http.StatusForbidden, nil
			}
			// DAV files are untrusted downloads on the application's origin.
			// Prevent uploaded HTML/SVG/JS from becoming executable UI content.
			w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'; script-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": path.Base(src)}))
		case "PUT":
			_, err := d.user.Fs.Stat(src)
			if (err == nil && !p.Modify) || (os.IsNotExist(err) && !p.Create) {
				return http.StatusForbidden, nil
			}
		case "MKCOL":
			if !p.Create {
				return http.StatusForbidden, nil
			}
		case "DELETE":
			if !p.Delete || src == "/" {
				return http.StatusForbidden, nil
			}
			if err := checkDescendants(d, src, ""); err != nil {
				return errToStatus(err), err
			}
		case "COPY", "MOVE":
			if (r.Method == "COPY" && (!p.Download || !p.Create)) || (r.Method == "MOVE" && !p.Rename) {
				return http.StatusForbidden, nil
			}
			u, err := url.Parse(r.Header.Get("Destination"))
			if err != nil || (u.Host != "" && u.Host != r.Host) || !strings.HasPrefix(u.Path, prefix+"/") {
				return http.StatusBadRequest, nil
			}
			dst := slashClean(strings.TrimPrefix(u.Path, prefix))
			if src == "/" || dst == "/" || dst == src || strings.HasPrefix(dst, src+"/") || strings.HasPrefix(src, dst+"/") || !d.Check(dst) {
				return http.StatusForbidden, nil
			}
			if err := checkDescendants(d, src, dst); err != nil {
				return errToStatus(err), err
			}
			if _, err := d.user.Fs.Stat(dst); err == nil {
				if !p.Delete {
					return http.StatusForbidden, nil
				}
				if err := checkDescendants(d, dst, ""); err != nil {
					return errToStatus(err), err
				}
			}
		case "LOCK", "UNLOCK":
			if !p.Create && !p.Modify {
				return http.StatusForbidden, nil
			}
		case "PROPPATCH":
			if !p.Modify {
				return http.StatusForbidden, nil
			}
		case "OPTIONS", "PROPFIND":
		default:
			return http.StatusMethodNotAllowed, nil
		}
		lockSource := src
		if r.Method == "COPY" {
			// COPY reads its source and only locks the destination.
			lockSource = ""
		}
		h := webdav.Handler{Prefix: prefix, FileSystem: &davFS{d: d, cache: cache}, LockSystem: davLocks{locks: locks, d: d, source: &lockSource}}
		h.ServeHTTP(w, r)
		return 0, nil
	}
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.settings.WebDAVEnabled(d.server) {
			return http.StatusNotFound, nil
		}
		if r.Header.Get("X-Auth") != "" {
			return withUser(serve)(w, r, d)
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="FileNeko WebDAV", charset="UTF-8"`)
		username, password, ok := r.BasicAuth()
		if !ok || d.settings.AuthMethod != fbAuth.MethodJSONAuth {
			return http.StatusUnauthorized, nil
		}
		a, err := d.store.Auth.Get(fbAuth.MethodJSONAuth)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		j, ok := a.(*fbAuth.JSONAuth)
		if !ok || (j.ReCaptcha != nil && j.ReCaptcha.Secret != "") {
			return http.StatusUnauthorized, nil
		}
		credentials, _ := json.Marshal(map[string]string{"username": username, "password": password})
		login := r.Clone(r.Context())
		login.Body = io.NopCloser(bytes.NewReader(credentials))
		d.user, err = j.Auth(login, d.store.Users, d.settings, d.server)
		if errors.Is(err, fberrors.ErrTooManyAttempts) {
			setPasswordRetryAfter(w, err)
			return http.StatusTooManyRequests, nil
		}
		if err != nil {
			return http.StatusUnauthorized, nil
		}
		w.Header().Del("WWW-Authenticate")
		return serve(w, r, d)
	}
}

// davFS retains the scoped, symlink-aware filesystem and applies rules to every
// operation, including the recursive calls made by the WebDAV implementation.
type davFS struct {
	d     *data
	cache FileCache
}

// Preview keys include modification times with second precision. Invalidate
// before mutation so a second write in the same second cannot serve old images.
func (f *davFS) invalidate(ctx context.Context, name string, recursive bool) error {
	if f.cache == nil {
		return nil
	}
	invalidateFile := func(name string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file := &files.FileInfo{Fs: f.d.user.Fs, Path: name, ModTime: info.ModTime()}
		return delThumbs(ctx, f.cache, file)
	}
	if recursive {
		return afero.Walk(f.d.user.Fs, name, invalidateFile)
	}
	info, err := f.d.user.Fs.Stat(name)
	if os.IsNotExist(err) {
		return nil
	}
	return invalidateFile(name, info, err)
}

func (f *davFS) Stat(_ context.Context, name string) (os.FileInfo, error) {
	name = slashClean(name)
	if !f.d.Check(name) {
		return nil, os.ErrPermission
	}
	info, err := f.d.user.Fs.Stat(name)
	if err == nil && !info.IsDir() && !info.Mode().IsRegular() {
		return nil, os.ErrPermission
	}
	return info, err
}

func (f *davFS) Mkdir(_ context.Context, name string, _ os.FileMode) error {
	name = slashClean(name)
	if !f.d.Check(name) || !f.d.user.Perm.Create {
		return os.ErrPermission
	}
	mode := f.d.settings.DirMode
	if mode == 0 {
		mode = settings.DefaultDirMode
	}
	return f.d.user.Fs.Mkdir(name, mode)
}

func (f *davFS) OpenFile(ctx context.Context, name string, flag int, _ os.FileMode) (webdav.File, error) {
	name = slashClean(name)
	// Opening a FIFO can block before a descriptor exists to inspect. Stat
	// first, for both reads and writes, and never open a special file.
	info, err := f.Stat(ctx, name)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if flag&(os.O_WRONLY|os.O_RDWR|os.O_CREATE|os.O_TRUNC) != 0 {
		if err == nil && info.IsDir() {
			return nil, os.ErrPermission
		}
		if (err == nil && !f.d.user.Perm.Modify) || (os.IsNotExist(err) && !f.d.user.Perm.Create) {
			return nil, os.ErrPermission
		}
		// A create-only account must not overwrite a file created between Stat
		// and OpenFile by another request.
		if os.IsNotExist(err) && !f.d.user.Perm.Modify {
			flag |= os.O_EXCL
		}
		if err := f.invalidate(ctx, name, false); err != nil {
			return nil, err
		}
	}
	mode := f.d.settings.FileMode
	if mode == 0 {
		mode = settings.DefaultFileMode
	}
	file, err := f.d.user.Fs.OpenFile(name, flag, mode)
	if err != nil {
		return nil, err
	}
	return &davFile{File: file, fs: f, name: name}, nil
}

func (f *davFS) RemoveAll(ctx context.Context, name string) error {
	name = slashClean(name)
	if name == "/" || !f.d.user.Perm.Delete || !f.d.Check(name) {
		return os.ErrPermission
	}
	if err := checkDescendants(f.d, name, ""); err != nil {
		return err
	}
	if err := f.invalidate(ctx, name, true); err != nil {
		return err
	}
	if err := revokeSharesForPath(f.d, name); err != nil {
		return err
	}
	return f.d.user.Fs.RemoveAll(name)
}

func (f *davFS) Rename(ctx context.Context, oldName, newName string) error {
	oldName, newName = slashClean(oldName), slashClean(newName)
	if oldName == "/" || newName == "/" || !f.d.user.Perm.Rename || !f.d.Check(oldName) || !f.d.Check(newName) {
		return os.ErrPermission
	}
	if err := checkDescendants(f.d, oldName, newName); err != nil {
		return err
	}
	if err := f.invalidate(ctx, oldName, true); err != nil {
		return err
	}
	if err := revokeSharesForPath(f.d, oldName); err != nil {
		return err
	}
	return f.d.user.Fs.Rename(oldName, newName)
}

type davFile struct {
	afero.File
	fs   *davFS
	name string
}

func (f *davFile) Readdir(n int) ([]os.FileInfo, error) {
	var visible []os.FileInfo
	for {
		count := n
		if n > 0 {
			count = n - len(visible)
		}
		entries, err := f.File.Readdir(count)
		for _, entry := range entries {
			if _, e := f.fs.Stat(context.Background(), path.Join(f.name, entry.Name())); e == nil {
				visible = append(visible, entry)
			}
		}
		if err != nil || n <= 0 || len(visible) >= n {
			return visible, err
		}
	}
}

// LockDetails roots are stored as physical paths so locks work across accounts
// with overlapping scopes. Refresh converts them back for DAV response hrefs.
type davLocks struct {
	locks  *davLockRegistry
	d      *data
	source *string
}

func (l davLocks) Confirm(now time.Time, name0, name1 string, conditions ...webdav.Condition) (func(), error) {
	if l.source != nil {
		// x/net may take name0 from a client-controlled tagged If URI. A
		// condition on another resource must not authorize this request's write.
		name0 = *l.source
	}
	var err error
	if name0 != "" {
		name0, err = l.physicalPath(name0)
		if err != nil {
			return nil, err
		}
	}
	if name1 != "" {
		name1, err = l.physicalPath(name1)
		if err != nil {
			return nil, err
		}
	}
	return l.locks.Confirm(now, name0, name1, conditions...)
}

func (l davLocks) Create(now time.Time, details webdav.LockDetails) (string, error) {
	virtual := slashClean(details.Root)
	var err error
	details.Root, err = l.physicalPath(details.Root)
	if err != nil {
		return "", err
	}
	return l.locks.Create(now, details, l.d.user.ID, virtual)
}

func (l davLocks) physicalPath(name string) (string, error) {
	resolved, err := files.ResolvePath(l.d.user.FullPath(slashClean(name)))
	if err != nil {
		return "", err
	}
	// memLS uses slash-separated, absolute keys on every platform.
	key := "/" + strings.TrimPrefix(filepath.ToSlash(resolved), "/")
	if l.d.server.CaseInsensitiveFs {
		key = strings.ToLower(key)
	}
	return key, nil
}

func (l davLocks) Refresh(now time.Time, token string, duration time.Duration) (webdav.LockDetails, error) {
	return l.locks.Refresh(now, token, duration, l.virtualEntry)
}

func (l davLocks) virtualEntry(entry davLockEntry) (string, error) {
	// Retain a creator's authorized alias, including when the administrator
	// explicitly permits external symlinks. Do not trust it after retargeting.
	if entry.owner == l.d.user.ID && l.d.Check(entry.virtual) {
		physical, err := l.physicalPath(entry.virtual)
		if err == nil && physical == entry.root {
			return entry.virtual, nil
		}
	}
	return l.virtualLockPath(entry.root)
}

func (l davLocks) virtualLockPath(physical string) (string, error) {
	root, err := l.physicalPath("/")
	if err != nil || (physical != root && !strings.HasPrefix(physical, strings.TrimSuffix(root, "/")+"/")) {
		return "", webdav.ErrForbidden
	}
	name := slashClean(strings.TrimPrefix(physical, root))
	if !l.d.Check(name) {
		return "", webdav.ErrForbidden
	}
	return name, nil
}

func (l davLocks) allowsLockRoot(entry davLockEntry) bool {
	_, err := l.virtualEntry(entry)
	return err == nil
}

func (l davLocks) Unlock(now time.Time, token string) error {
	return l.locks.Unlock(now, token, l.allowsLockRoot)
}
