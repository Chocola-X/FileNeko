package fbhttp

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/share"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
	"golang.org/x/crypto/bcrypt"
)

func proxyAuthFixture(t *testing.T, server *settings.Server) (http.Handler, *storage.Storage, string) {
	t.Helper()
	root := t.TempDir()
	key := []byte("trusted-proxy-fixture")
	st := scopedUserStorage(t, root, users.Permissions{Download: true}, key)
	user, err := st.Users.Get(root, false, "u")
	if err != nil {
		t.Fatal(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("proxy-password"), bcrypt.MinCost)
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
	if err := st.Settings.Save(&settings.Settings{Key: key, AuthMethod: auth.MethodJSONAuth,
		Fail2Ban: &settings.Fail2Ban{Enabled: true, MaxAttempts: 3, WindowMinutes: 5, BanMinutes: 5}}); err != nil {
		t.Fatal(err)
	}
	server.Root = root
	server.EnableWebDAV = true
	h, err := NewHandler(nil, nil, nil, st, server, os.DirFS(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	return h, st, string(hash)
}

func TestTrustedProxySeparatesClientsAndUnifiesLoginDAVAndLogs(t *testing.T) {
	server := &settings.Server{TrustedProxies: []string{"192.0.2.201/32"}}
	h, _, _ := proxyAuthFixture(t, server)
	var logs bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(previous)
	request := func(dav bool, remote, forwarded, password string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"username":"u","password":"`+password+`"}`))
		if dav {
			r = httptest.NewRequest("PROPFIND", "/dav/", nil)
			r.SetBasicAuth("u", password)
			r.Header.Set("Depth", "0")
		}
		r.RemoteAddr = remote
		r.Header.Set("X-Forwarded-For", forwarded)
		r.Header.Set("X-Real-IP", "203.0.113.200")
		r.Header.Set("X-Forwarded-Proto", "https")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	for i, dav := range []bool{false, true, false} {
		w := request(dav, "192.0.2.201:1234", "203.0.113.99, 198.19.1.20", "incorrect")
		want := 403
		if dav {
			want = 401
		}
		if w.Code != want {
			t.Fatalf("attempt %d: %d", i, w.Code)
		}
	}
	for _, dav := range []bool{false, true} {
		// Changing the forged prefix and port cannot escape the client's ban.
		w := request(dav, "192.0.2.201:9999", "203.0.113.100, 198.19.1.20", "proxy-password")
		if w.Code != 429 {
			t.Fatalf("client escaped ban: %d", w.Code)
		}
	}
	if w := request(false, "192.0.2.201:1234", "198.19.1.21", "proxy-password"); w.Code != 200 {
		t.Fatalf("another client behind the same proxy was blocked: %d", w.Code)
	}
	// A direct caller cannot borrow the trusted proxy's header privileges.
	for i := 0; i < 3; i++ {
		if w := request(false, "198.19.1.22:1234", "198.19.2."+string(rune('1'+i)), "incorrect"); w.Code != 403 {
			t.Fatalf("direct attempt %d: %d", i, w.Code)
		}
	}
	if w := request(true, "198.19.1.22:9999", "198.19.2.100", "proxy-password"); w.Code != 429 {
		t.Fatalf("spoofed direct source escaped ban: %d", w.Code)
	}
	output := logs.String()
	if !strings.Contains(output, "403 198.19.1.20") || !strings.Contains(output, "401 198.19.1.20") || !strings.Contains(output, "429 198.19.1.22") {
		t.Fatalf("logs do not use the same resolved source as bans: %s", output)
	}
	if strings.Contains(output, "203.0.113.") || strings.Contains(output, "192.0.2.201") {
		t.Fatalf("logs trusted forged metadata or reported proxy as client: %s", output)
	}
}

func TestSecureCookiesBehindTrustedTLSProxy(t *testing.T) {
	for _, tc := range []struct {
		name, remote, scheme, proto string
		trusted                     []string
		secure                      bool
	}{
		{"trusted TLS termination", "192.0.2.202:1234", "http", "https", []string{"192.0.2.202"}, true},
		{"untrusted TLS claim", "198.19.1.72:1234", "http", "https", []string{"192.0.2.202"}, false},
		{"forwarding disabled", "192.0.2.202:1234", "http", "https", nil, false},
		{"native TLS cannot downgrade", "198.19.1.73:1234", "https", "http", nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := &settings.Server{TrustedProxies: tc.trusted}
			h, st, hash := proxyAuthFixture(t, server)
			request := func(path, token string) *httptest.ResponseRecorder {
				r := httptest.NewRequest("POST", tc.scheme+"://example.com"+path, strings.NewReader(`{"username":"u","password":"proxy-password"}`))
				r.RemoteAddr = tc.remote
				r.Header.Set("X-Forwarded-For", "198.19.1.71")
				r.Header.Set("X-Forwarded-Proto", tc.proto)
				if token != "" {
					r.Header.Set("X-Auth", token)
				}
				w := httptest.NewRecorder()
				h.ServeHTTP(w, r)
				if w.Code != 200 && w.Code != 204 {
					t.Fatalf("%s: %d %s", path, w.Code, w.Body)
				}
				cookies := w.Result().Cookies()
				if len(cookies) != 1 || cookies[0].Secure != tc.secure || !cookies[0].HttpOnly {
					t.Fatalf("%s cookies: %+v", path, cookies)
				}
				return w
			}
			login := request("/api/login", "")
			renewal := request("/api/renew", login.Body.String())
			request("/api/logout", renewal.Body.String())
			r := httptest.NewRequest("GET", tc.scheme+"://example.com/api/public/share/h", nil)
			r.RemoteAddr = tc.remote
			r.Header.Set("X-Forwarded-Proto", tc.proto)
			r.Header.Set("X-SHARE-PASSWORD", "proxy-password")
			set, err := st.Settings.Get()
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			status, err := authenticateShareRequest(w, r, &data{server: server, settings: set}, &share.Link{Hash: "h", PasswordHash: hash, Token: "binding"})
			if err != nil || status != 0 {
				t.Fatalf("share authentication: %d %v", status, err)
			}
			cookies := w.Result().Cookies()
			if len(cookies) != 1 || cookies[0].Secure != tc.secure || !cookies[0].HttpOnly {
				t.Fatalf("share cookies: %+v", cookies)
			}
		})
	}
}

func TestInvalidTrustedProxyConfigRejectsHandlerStartup(t *testing.T) {
	if _, err := NewHandler(nil, nil, nil, nil, &settings.Server{TrustedProxies: []string{"localhost"}}, nil); err == nil {
		t.Fatal("invalid trust config was accepted at startup")
	}
}
