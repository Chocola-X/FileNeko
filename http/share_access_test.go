package fbhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/share"
)

// GHSA-3v48-283x-f2w4 / #5239: URLs must not carry a permanent password
// bypass. Browser downloads instead reuse a bounded, share-specific cookie.
func TestPasswordShareAccessIsNotGrantedByURL(t *testing.T) {
	link := &share.Link{Hash: "h", PasswordHash: "$2y$10$TFAmdCbyd/mEZDe5fUeZJu.MaJQXRTwdqb/IQV.eTn6dWrF58gCSe", Token: "legacy-secret"}
	d := &data{settings: &settings.Settings{Key: []byte("key")}, server: &settings.Server{BaseURL: "/files"}}
	req := httptest.NewRequest("GET", "https://example.com/files/api/public/share/h?token=legacy-secret", nil)
	rec := httptest.NewRecorder()
	if status, err := authenticateShareRequest(rec, req, d, link); err != nil || status != 401 {
		t.Fatalf("URL token accepted: %d %v", status, err)
	}
	req.Header.Set("X-SHARE-PASSWORD", "password")
	if status, err := authenticateShareRequest(rec, req, d, link); err != nil || status != 0 {
		t.Fatalf("password rejected: %d %v", status, err)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].Path != "/files/api/public/" || cookies[0].MaxAge > int(shareAccessTTL.Seconds()) {
		t.Fatalf("unsafe share cookie: %+v", cookies)
	}
	check := func(cookie *http.Cookie) int {
		req := httptest.NewRequest("GET", "/files/api/public/dl/h/file.txt", nil)
		req.AddCookie(cookie)
		status, _ := authenticateShareRequest(httptest.NewRecorder(), req, d, link)
		return status
	}
	cookie := cookies[0]
	if status := check(cookie); status != 0 {
		t.Fatalf("native download cookie: %d", status)
	}
	saved := cookie.Value
	cookie.Value = shareAccessValue(d.settings.Key, link, time.Now().Add(-time.Second).Unix())
	if status := check(cookie); status != 401 {
		t.Fatalf("expired access: %d", status)
	}
	cookie.Value = saved + "tampered"
	if status := check(cookie); status != 401 {
		t.Fatalf("tampered access: %d", status)
	}
	cookie.Value = saved
	link.Token = "rotated"
	if status := check(cookie); status != 401 {
		t.Fatalf("old share credentials survived rotation: %d", status)
	}
	link.Token = "legacy-secret"
	link.Hash = "another-share"
	if status := check(cookie); status != 401 {
		t.Fatalf("cookie used for another share: %d", status)
	}
}
