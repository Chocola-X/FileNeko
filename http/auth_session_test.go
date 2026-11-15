package fbhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/golang-jwt/jwt/v5"
)

// GHSA-7xwp-2cpp-p8r7: a correctly signed token is no longer enough once its
// server-side session has been revoked. Missing IDs from old releases fail shut.
func TestLogoutRevokesSession(t *testing.T) {
	key := []byte("key")
	perm := users.Permissions{}
	st := scopedUserStorage(t, t.TempDir(), perm, key)
	token := signToken(t, perm, key, st)
	protected := withUser(func(http.ResponseWriter, *http.Request, *data) (int, error) { return 204, nil })
	call := func(handler handleFunc, token string) int {
		req := httptest.NewRequest("POST", "/", nil)
		req.Header.Set("X-Auth", token)
		rec := httptest.NewRecorder()
		handle(handler, "", st, &settings.Server{}).ServeHTTP(rec, req)
		return rec.Code
	}
	if got := call(protected, token); got != 204 {
		t.Fatalf("active session: %d", got)
	}
	if got := call(logoutHandler, token); got != 204 {
		t.Fatalf("logout: %d", got)
	}
	if got := call(protected, token); got != 401 {
		t.Fatalf("revoked session: %d", got)
	}
	if got := call(renewHandler(DefaultTokenExpirationTime), token); got != 401 {
		t.Fatalf("revoked session renewed: %d", got)
	}
	var claims authToken
	if _, err := jwt.ParseWithClaims(token, &claims, func(*jwt.Token) (interface{}, error) { return key, nil }); err != nil {
		t.Fatal(err)
	}
	claims.ID = ""
	legacy, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims).SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	if got := call(protected, legacy); got != 401 {
		t.Fatalf("legacy stateless token accepted: %d", got)
	}
}

func TestLoginSetsProtectedCookie(t *testing.T) {
	st := scopedUserStorage(t, t.TempDir(), users.Permissions{}, []byte("key"))
	user, err := st.Users.Get("", false, uint(1))
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "https://example.com/files/api/login", nil)
	d := &data{store: st, settings: &settings.Settings{Key: []byte("key")}, server: &settings.Server{BaseURL: "/files"}}
	if _, err := printToken(rec, req, d, user, DefaultTokenExpirationTime); err != nil {
		t.Fatal(err)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].SameSite != http.SameSiteStrictMode || cookies[0].Path != "/files/" {
		t.Fatalf("unsafe session cookie: %+v", cookies)
	}
}

func TestRenewConsumesOldToken(t *testing.T) {
	key := []byte("key")
	perm := users.Permissions{}
	st := scopedUserStorage(t, t.TempDir(), perm, key)
	old := signToken(t, perm, key, st)
	renew := func(token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", "/renew", nil)
		req.Header.Set("X-Auth", token)
		rec := httptest.NewRecorder()
		handle(renewHandler(DefaultTokenExpirationTime), "", st, &settings.Server{}).ServeHTTP(rec, req)
		return rec
	}
	first := renew(old)
	if first.Code != 200 {
		t.Fatalf("renew: %d %s", first.Code, first.Body)
	}
	if next := renew(old); next.Code != 401 {
		t.Fatalf("token redeemed twice: %d", next.Code)
	}
	if next := renew(first.Body.String()); next.Code != 200 {
		t.Fatalf("successor rejected: %d", next.Code)
	}
}

// Logout must revoke the same login's successor even if renewal wins the race
// and the browser sends logout with the just-consumed identifier.
func TestLogoutRevokesRenewedSessionFamily(t *testing.T) {
	key := []byte("key")
	perm := users.Permissions{}
	st := scopedUserStorage(t, t.TempDir(), perm, key)
	old := signToken(t, perm, key, st)
	call := func(handler handleFunc, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", "/", nil)
		req.Header.Set("X-Auth", token)
		rec := httptest.NewRecorder()
		handle(handler, "", st, &settings.Server{}).ServeHTTP(rec, req)
		return rec
	}
	renewed := call(renewHandler(DefaultTokenExpirationTime), old)
	if renewed.Code != 200 {
		t.Fatalf("renew: %d", renewed.Code)
	}
	if rec := call(logoutHandler, old); rec.Code != 204 {
		t.Fatalf("logout after rotation: %d", rec.Code)
	}
	protected := withUser(func(http.ResponseWriter, *http.Request, *data) (int, error) { return 204, nil })
	if rec := call(protected, renewed.Body.String()); rec.Code != 401 {
		t.Fatalf("successor survived logout: %d", rec.Code)
	}
}
