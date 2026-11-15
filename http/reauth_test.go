package fbhttp

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/gorilla/mux"
	"strings"
)

// #5213: omitting which, changing its case, or changing hook settings cannot
// bypass the current-password requirement on sensitive mutations.
func TestSensitiveUpdatesRequireCurrentPassword(t *testing.T) {
	key := []byte("key")
	perm := users.Permissions{Admin: true}
	st := scopedUserStorage(t, t.TempDir(), perm, key)
	user, err := st.Users.Get("", false, uint(1))
	if err != nil {
		t.Fatal(err)
	}
	user.Password, err = users.HashPwd("current-password")
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Users.Update(user, "Password"); err != nil {
		t.Fatal(err)
	}
	if err := st.Settings.Save(&settings.Settings{Key: key, AuthMethod: auth.MethodJSONAuth}); err != nil {
		t.Fatal(err)
	}
	token := signToken(t, perm, key, st)
	for _, which := range [][]string{nil, {}, {"all"}, {"LOCKPASSWORD"}, {"rules"}} {
		body, err := json.Marshal(map[string]interface{}{"what": "user", "which": which, "data": user})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("PUT", "/users/1", strings.NewReader(string(body)))
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		req.Header.Set("X-Auth", token)
		rec := httptest.NewRecorder()
		handle(userPutHandler, "", st, &settings.Server{}).ServeHTTP(rec, req)
		if rec.Code != 400 {
			t.Errorf("which=%v without password: %d %s", which, rec.Code, rec.Body)
		}
	}
	for _, password := range []string{"", "wrong", "current-password"} {
		body, err := json.Marshal(map[string]interface{}{"current_password": password, "commands": map[string][]string{"after_upload": {"echo test"}}})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("PUT", "/settings", strings.NewReader(string(body)))
		req.Header.Set("X-Auth", token)
		rec := httptest.NewRecorder()
		handle(settingsPutHandler, "", st, &settings.Server{}).ServeHTTP(rec, req)
		want := 400
		if password == "current-password" {
			want = 200
		}
		if rec.Code != want {
			t.Errorf("settings password=%q: %d want %d", password, rec.Code, want)
		}
	}
}
