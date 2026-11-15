package fbhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// GHSA-v7vv-5wj2-gfcj: password changes invalidate old tokens immediately,
// even within the same second and after re-opening persistent user storage.
func TestPasswordChangeInvalidatesToken(t *testing.T) {
	key := []byte("key")
	perm := users.Permissions{}
	st := scopedUserStorage(t, t.TempDir(), perm, key)
	old := signToken(t, perm, key, st)
	protected := handle(withUser(func(http.ResponseWriter, *http.Request, *data) (int, error) { return 204, nil }), "", st, &settings.Server{})
	get := func(token string) int {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Auth", token)
		rec := httptest.NewRecorder()
		protected.ServeHTTP(rec, req)
		return rec.Code
	}
	if got := get(old); got != 204 {
		t.Fatalf("initial token: %d", got)
	}
	user, err := st.Users.Get("", false, uint(1))
	if err != nil {
		t.Fatal(err)
	}
	user.Password = "changed-hash"
	if err := st.Users.Update(user, "Password"); err != nil {
		t.Fatal(err)
	}
	if got := get(old); got != 401 {
		t.Fatalf("old token after password change: %d", got)
	}
	req := httptest.NewRequest("POST", "/renew", nil)
	req.Header.Set("X-Auth", old)
	rec := httptest.NewRecorder()
	handle(renewHandler(DefaultTokenExpirationTime), "", st, &settings.Server{}).ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("old token renewed: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	d := &data{settings: &settings.Settings{Key: key}, store: st}
	if _, err := printToken(rec, req, d, user, DefaultTokenExpirationTime); err != nil {
		t.Fatal(err)
	}
	if got := get(rec.Body.String()); got != 204 {
		t.Fatalf("new token rejected: %d", got)
	}
}
