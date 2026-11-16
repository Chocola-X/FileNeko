package fbhttp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

func TestAdminNewUserHome(t *testing.T) {
	for _, tc := range []struct {
		name, scope, base, want string
		auto                    bool
	}{
		{name: "omitted scope", want: "/users/alice"},
		{name: "blank scope", scope: "  ", want: "/users/alice"},
		{name: "auto setting enabled", auto: true, want: "/users/alice"},
		{name: "custom home base", base: "/readers", want: "/readers/alice"},
		{name: "explicit shared scope", scope: "/shared", want: "/shared"},
		{name: "explicit root scope", scope: "/", want: "/"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "secret.txt"), []byte("root only"), 0600); err != nil {
				t.Fatal(err)
			}
			key := []byte("test-signing-key")
			perm := users.Permissions{Admin: true}
			st := scopedUserStorage(t, root, perm, key)
			admin, err := st.Users.Get(root, false, uint(1))
			if err != nil {
				t.Fatal(err)
			}
			admin.Password, err = users.HashPwd("CurrentPassword54321!")
			if err != nil {
				t.Fatal(err)
			}
			if err = st.Users.Update(admin, "Password"); err != nil {
				t.Fatal(err)
			}
			if err = st.Settings.Save(&settings.Settings{
				Key: key, AuthMethod: auth.MethodJSONAuth, CreateUserDir: tc.auto,
				UserHomeBasePath: tc.base, Defaults: settings.UserDefaults{Scope: "/"},
			}); err != nil {
				t.Fatal(err)
			}
			token := signToken(t, perm, key, st)
			create := func(name, scope string) *httptest.ResponseRecorder {
				body, marshalErr := json.Marshal(map[string]interface{}{
					"what": "user", "current_password": "CurrentPassword54321!",
					"data": users.User{Username: name, Password: "NewPassword54321!", Scope: scope},
				})
				if marshalErr != nil {
					t.Fatal(marshalErr)
				}
				req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
				req.Header.Set("X-Auth", token)
				rec := httptest.NewRecorder()
				handle(userPostHandler, "", st, &settings.Server{Root: root}).ServeHTTP(rec, req)
				return rec
			}
			if rec := create("alice", tc.scope); rec.Code != http.StatusCreated {
				t.Fatalf("create: %d %s", rec.Code, rec.Body)
			}
			user, err := st.Users.GetByScope(tc.want)
			if err != nil || user == nil || user.Username != "alice" {
				t.Fatalf("expected alice at %s, got %v, %v", tc.want, user, err)
			}
			if _, err = os.Stat(filepath.Join(root, tc.want)); err != nil {
				t.Fatal(err)
			}
			if tc.scope == "" || tc.scope == "  " {
				if err = user.Clean(root, false); err != nil {
					t.Fatal(err)
				}
				if _, err = user.Fs.Open("/secret.txt"); !os.IsNotExist(err) {
					t.Fatalf("new user can access root secret: %v", err)
				}
				// A different username normalizing to an existing home cannot
				// silently acquire that home through automatic allocation.
				if rec := create("al/ice", ""); rec.Code != http.StatusCreated {
					t.Fatalf("second user: %d %s", rec.Code, rec.Body)
				}
				if rec := create("al-ice", ""); rec.Code != http.StatusConflict {
					t.Fatalf("colliding home: %d %s", rec.Code, rec.Body)
				}
			}
		})
	}
}
