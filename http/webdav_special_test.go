package fbhttp

import (
	"context"
	"os"
	"testing"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/afero"
)

type davSpecialFs struct{ specialFs }

func (f davSpecialFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if name == "/special" {
		panic("WebDAV must not open a pipe, device or socket")
	}
	return f.Fs.OpenFile(name, flag, perm)
}

func TestWebDAVRejectsSpecialFilesBeforeOpening(t *testing.T) {
	for _, mode := range []os.FileMode{os.ModeNamedPipe, os.ModeDevice, os.ModeSocket} {
		t.Run(mode.String(), func(t *testing.T) {
			base := afero.NewMemMapFs()
			if err := afero.WriteFile(base, "/special", nil, 0600); err != nil {
				t.Fatal(err)
			}
			fs := davSpecialFs{specialFs{Fs: base, mode: mode}}
			key := []byte("dav-special-file")
			perms := users.Permissions{Download: true, Create: true, Modify: true}
			st := scopedUserStorage(t, t.TempDir(), perms, key)
			st.Users = &customFSUser{Store: st.Users, fs: fs}
			token := signToken(t, perms, key, st)
			h := handle(newWebDAVHandler(nil), "", st, &settings.Server{EnableWebDAV: true})
			for _, method := range []string{"GET", "HEAD", "PUT", "COPY", "PROPFIND"} {
				w := davSecurityRequest(h, token, method, "/special", "", map[string]string{"Destination": "/dav/copy", "Depth": "0"})
				if w.Code < 400 {
					t.Fatalf("%s allowed special file: %d", method, w.Code)
				}
			}
			// Directory listings must also omit unusable special entries.
			d := &data{settings: &settings.Settings{}, server: &settings.Server{}, user: &users.User{Fs: fs, Perm: perms}}
			f := &davFS{d: d}
			dir, err := f.OpenFile(context.Background(), "/", os.O_RDONLY, 0)
			if err != nil {
				t.Fatal(err)
			}
			defer dir.Close()
			entries, err := dir.Readdir(-1)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Fatal("special file advertised in directory listing")
			}
		})
	}
}
