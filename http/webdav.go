package fbhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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
	locks := webdav.NewMemLS()
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
		h := webdav.Handler{Prefix: prefix, FileSystem: &davFS{d: d, cache: cache}, LockSystem: davLocks{LockSystem: locks, d: d}}
		h.ServeHTTP(w, r)
		return 0, nil
	}
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if r.Header.Get("X-Auth") != "" {
			return withUser(serve)(w, r, d)
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="File Browser WebDAV", charset="UTF-8"`)
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
	clear := func(name string, info os.FileInfo, err error) error {
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
		return afero.Walk(f.d.user.Fs, name, clear)
	}
	info, err := f.d.user.Fs.Stat(name)
	if os.IsNotExist(err) {
		return nil
	}
	return clear(name, info, err)
}

func (f *davFS) Stat(_ context.Context, name string) (os.FileInfo, error) {
	name = slashClean(name)
	if !f.d.Check(name) {
		return nil, os.ErrPermission
	}
	return f.d.user.Fs.Stat(name)
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
	if !f.d.Check(name) {
		return nil, os.ErrPermission
	}
	if flag&(os.O_WRONLY|os.O_RDWR|os.O_CREATE|os.O_TRUNC) != 0 {
		_, err := f.d.user.Fs.Stat(name)
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
	webdav.LockSystem
	d *data
}

func (l davLocks) Confirm(now time.Time, name0, name1 string, conditions ...webdav.Condition) (func(), error) {
	if name0 != "" {
		name0 = l.d.user.FullPath(slashClean(name0))
	}
	if name1 != "" {
		name1 = l.d.user.FullPath(slashClean(name1))
	}
	return l.LockSystem.Confirm(now, name0, name1, conditions...)
}

func (l davLocks) Create(now time.Time, details webdav.LockDetails) (string, error) {
	details.Root = l.d.user.FullPath(slashClean(details.Root))
	return l.LockSystem.Create(now, details)
}

func (l davLocks) Refresh(now time.Time, token string, duration time.Duration) (webdav.LockDetails, error) {
	details, err := l.LockSystem.Refresh(now, token, duration)
	if err == nil {
		rel, e := filepath.Rel(l.d.user.FullPath("/"), details.Root)
		if e != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return webdav.LockDetails{}, os.ErrPermission
		}
		details.Root = slashClean(rel)
	}
	return details, err
}
