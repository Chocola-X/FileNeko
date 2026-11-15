package fbhttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
	"golang.org/x/crypto/bcrypt"
)

func TestSecuritySettingsPersistAndWebDAVToggleImmediately(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "demo.txt"), []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	key := []byte("security-settings-test")
	perm := users.Permissions{Admin: true, Download: true}
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
	server := &settings.Server{Root: root}
	h, err := NewHandler(nil, nil, nil, st, server, os.DirFS(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	token := signToken(t, perm, key, st)
	request := func(method, path string, body any, session bool) *httptest.ResponseRecorder {
		t.Helper()
		payload := ""
		if body != nil {
			b, e := json.Marshal(body)
			if e != nil {
				t.Fatal(e)
			}
			payload = string(b)
		}
		r := httptest.NewRequest(method, path, strings.NewReader(payload))
		r.RemoteAddr = "198.19.0.20:1234"
		if session {
			r.Header.Set("X-Auth", token)
		} else {
			r.SetBasicAuth("u", "settings-password")
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	assert := func(w *httptest.ResponseRecorder, status int) {
		t.Helper()
		if w.Code != status {
			t.Fatalf("status=%d want=%d: %s", w.Code, status, w.Body)
		}
	}
	w := request("GET", "/api/settings", nil, true)
	assert(w, 200)
	var dto settingsData
	if err := json.Unmarshal(w.Body.Bytes(), &dto); err != nil {
		t.Fatal(err)
	}
	if dto.Fail2Ban == nil || *dto.Fail2Ban != settings.DefaultFail2Ban() || dto.EnableWebDAV == nil || *dto.EnableWebDAV {
		t.Fatal("missing safe legacy defaults")
	}
	assert(request("GET", "/dav/demo.txt", nil, false), 404)
	dto.CurrentPassword = "settings-password"
	enabled := true
	dto.EnableWebDAV = &enabled
	dto.Fail2Ban = &settings.Fail2Ban{Enabled: true, MaxAttempts: 3, WindowMinutes: 2, BanMinutes: 5}
	assert(request("PUT", "/api/settings", dto, true), 200)
	assert(request("GET", "/dav/demo.txt", nil, false), 200)
	// New settings survive loading from the database, independently of server flags.
	persisted, err := st.Settings.Get()
	if err != nil {
		t.Fatal(err)
	}
	if !persisted.WebDAVEnabled(server) || *persisted.Fail2Ban != *dto.Fail2Ban {
		t.Fatal("security settings were not persisted")
	}
	dto.Fail2Ban.MaxAttempts = 0
	assert(request("PUT", "/api/settings", dto, true), 400)
	persisted, err = st.Settings.Get()
	if err != nil {
		t.Fatal(err)
	}
	if persisted.Fail2Ban.MaxAttempts != 3 {
		t.Fatal("invalid configuration was partially saved")
	}
	dto.Fail2Ban.MaxAttempts = 3
	enabled = false
	// An ordinary account cannot toggle DAV, even with the correct password.
	user.Perm.Admin = false
	if err := st.Users.Update(user, "Perm"); err != nil {
		t.Fatal(err)
	}
	assert(request("PUT", "/api/settings", dto, true), 403)
	user.Perm.Admin = true
	if err := st.Users.Update(user, "Perm"); err != nil {
		t.Fatal(err)
	}
	assert(request("PUT", "/api/settings", dto, true), 200)
	assert(request("GET", "/dav/demo.txt", nil, false), 404)
	assert(request("GET", "/dav/demo.txt", nil, true), 404)
	// A new router also observes the disabled value.
	h, err = NewHandler(nil, nil, nil, st, server, os.DirFS(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	assert(request("PROPFIND", "/dav/", nil, false), 404)
}

func TestConfiguredFail2BanIsSharedBetweenLoginAndWebDAV(t *testing.T) {
	root := t.TempDir()
	key := []byte("configured-ban-test")
	st := scopedUserStorage(t, root, users.Permissions{Download: true}, key)
	user, err := st.Users.Get(root, false, "u")
	if err != nil {
		t.Fatal(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
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
	enabled := true
	set := &settings.Settings{Key: key, AuthMethod: auth.MethodJSONAuth, EnableWebDAV: &enabled,
		Fail2Ban: &settings.Fail2Ban{Enabled: true, MaxAttempts: 3, WindowMinutes: 5, BanMinutes: 5}}
	if err := st.Settings.Save(set); err != nil {
		t.Fatal(err)
	}
	h, err := NewHandler(nil, nil, nil, st, &settings.Server{Root: root}, os.DirFS(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	request := func(dav bool, source, password string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"username":"u","password":"`+password+`"}`))
		if dav {
			r = httptest.NewRequest("PROPFIND", "/dav/", nil)
			r.SetBasicAuth("u", password)
			r.Header.Set("Depth", "0")
		}
		r.RemoteAddr = source
		r.Header.Set("X-Forwarded-For", "203.0.113."+strconv.Itoa(len(password)))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	// Correct requests can be frequent and never consume the failed-attempt threshold.
	for i := 0; i < 12; i++ {
		if w := request(true, "198.19.0.30:1000", "correct-password"); w.Code != 207 {
			t.Fatalf("DAV success counted as failure: %d", w.Code)
		}
	}
	for i, dav := range []bool{false, true, false} {
		w := request(dav, "198.19.0.30:1000", "wrong")
		want := http.StatusForbidden
		if dav {
			want = 401
		}
		if w.Code != want {
			t.Fatalf("attempt %d status=%d", i, w.Code)
		}
	}
	for _, dav := range []bool{false, true} {
		w := request(dav, "198.19.0.30:9999", "correct-password")
		retry, _ := strconv.Atoi(w.Header().Get("Retry-After"))
		if w.Code != 429 || retry < 295 || retry > 300 {
			t.Fatalf("ban missing: status=%d retry=%d", w.Code, retry)
		}
	}
	if w := request(false, "198.19.0.31:1000", "correct-password"); w.Code != 200 {
		t.Fatalf("independent source rejected: %d", w.Code)
	}
	// Turning the configurable ban off takes effect for both authentication entry points.
	set.Fail2Ban.Enabled = false
	if err := st.Settings.Save(set); err != nil {
		t.Fatal(err)
	}
	if w := request(true, "198.19.0.30:1000", "correct-password"); w.Code != 207 {
		t.Fatalf("disabled ban remained active: %d", w.Code)
	}
}
