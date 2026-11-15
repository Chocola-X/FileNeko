package fbhttp

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/diskcache"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginGuessLimitAlsoCoversWebDAV(t *testing.T) {
	key := []byte("key")
	st := scopedUserStorage(t, t.TempDir(), users.Permissions{Download: true}, key)
	user, err := st.Users.Get("", false, uint(1))
	if err != nil {
		t.Fatal(err)
	}
	hash, hashErr := bcrypt.GenerateFromPassword([]byte("valid-password-123"), bcrypt.MinCost)
	if hashErr != nil {
		t.Fatal(hashErr)
	}
	user.Password = string(hash)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Users.Update(user, "Password"); err != nil {
		t.Fatal(err)
	}
	if err := st.Auth.Save(&auth.JSONAuth{}); err != nil {
		t.Fatal(err)
	}
	if err := st.Settings.Save(&settings.Settings{Key: key, AuthMethod: auth.MethodJSONAuth}); err != nil {
		t.Fatal(err)
	}
	server := &settings.Server{EnableWebDAV: true}
	login := handle(loginHandler(DefaultTokenExpirationTime), "", st, server)
	var last *httptest.ResponseRecorder
	for i := 0; i < 11; i++ {
		req := httptest.NewRequest("POST", "/login", strings.NewReader(`{"username":"u","password":"wrong"}`))
		req.RemoteAddr = "198.18.0.10:1234"
		req.Header.Set("X-Forwarded-For", string(rune('a'+i)))
		last = httptest.NewRecorder()
		login.ServeHTTP(last, req)
	}
	if last.Code != 429 || last.Header().Get("Retry-After") == "" {
		t.Fatalf("unlimited guesses: %d %s", last.Code, last.Body)
	}
	req := httptest.NewRequest("GET", "/dav/", nil)
	req.RemoteAddr = "198.18.0.10:5678"
	req.SetBasicAuth("u", "wrong")
	rec := httptest.NewRecorder()
	handle(newWebDAVHandler(diskcache.NewNoOp()), "", st, server).ServeHTTP(rec, req)
	if rec.Code != 429 {
		t.Fatalf("WebDAV bypassed password limit: %d", rec.Code)
	}
	req = httptest.NewRequest("POST", "/login", strings.NewReader(`{"username":"u","password":"valid-password-123"}`))
	req.RemoteAddr = "198.18.0.11:1234"
	rec = httptest.NewRecorder()
	login.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("other peer cannot log in: %d %s", rec.Code, rec.Body)
	}
}
