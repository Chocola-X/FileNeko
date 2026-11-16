package fbhttp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/filebrowser/filebrowser/v2/version"
	"golang.org/x/crypto/bcrypt"
)

func TestHideVersionPersistsAndOmitsPublicBootstrapVersion(t *testing.T) {
	root := t.TempDir()
	key := []byte("branding-test")
	perm := users.Permissions{Admin: true}
	st := scopedUserStorage(t, root, perm, key)
	user, err := st.Users.Get(root, false, "u")
	if err != nil {
		t.Fatal(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("settings-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	user.Password = string(hash)
	if err := st.Users.Update(user, "Password"); err != nil {
		t.Fatal(err)
	}
	if err := st.Auth.Save(&auth.JSONAuth{}); err != nil {
		t.Fatal(err)
	}
	if err := st.Settings.Save(&settings.Settings{Key: key, AuthMethod: auth.MethodJSONAuth}); err != nil {
		t.Fatal(err)
	}
	assets := fstest.MapFS{"public/index.html": {Data: []byte("<script>[{[ .Json ]}]</script>")}}
	h, err := NewHandler(nil, nil, nil, st, &settings.Server{Root: root}, assets)
	if err != nil {
		t.Fatal(err)
	}
	token := signToken(t, perm, key, st)
	request := func(method, path string, body any) []byte {
		t.Helper()
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		r := httptest.NewRequest(method, path, bytes.NewReader(payload))
		if path == "/api/settings" {
			r.Header.Set("X-Auth", token)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body)
		}
		return w.Body.Bytes()
	}
	var dto settingsData
	if err := json.Unmarshal(request(http.MethodGet, "/api/settings", nil), &dto); err != nil {
		t.Fatal(err)
	}
	if dto.Branding.HideVersion {
		t.Fatal("legacy settings must show the version by default")
	}
	for _, hide := range []bool{false, true, false} {
		dto.CurrentPassword = "settings-password"
		dto.Branding.HideVersion = hide
		request(http.MethodPut, "/api/settings", dto)
		persisted, err := st.Settings.Get()
		if err != nil {
			t.Fatal(err)
		}
		if persisted.Branding.HideVersion != hide {
			t.Fatal("version visibility was not persisted")
		}
		for _, path := range []string{"/login", "/share/public", "/403", "/files/", "/settings/global"} {
			var bootstrap struct {
				Version     string
				HideVersion bool
				SourceURL   string
			}
			body := request(http.MethodGet, path, nil)
			body = bytes.TrimSuffix(bytes.TrimPrefix(body, []byte("<script>")), []byte("</script>"))
			if err := json.Unmarshal(body, &bootstrap); err != nil {
				t.Fatal(err)
			}
			wantVersion := version.Version
			if hide {
				wantVersion = ""
			}
			if bootstrap.Version != wantVersion || bootstrap.HideVersion != hide || bootstrap.SourceURL != version.SourceURL {
				t.Fatalf("%s hide=%t: unexpected bootstrap %+v", path, hide, bootstrap)
			}
		}
	}
}
