package fbhttp

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// GHSA-6759-996p-gpj6: enabling signup must not grant the server root or
// the administrator's shared default scope, even with CreateUserDir disabled.
func TestSignupAlwaysCreatesIsolatedNewHome(t *testing.T) {
	for _, defaultScope := range []string{"/", "/shared"} {
		t.Run(defaultScope, func(t *testing.T) {
			root := t.TempDir()
			key := []byte("key")
			if err := os.WriteFile(filepath.Join(root, "secret.txt"), []byte("private"), 0600); err != nil {
				t.Fatal(err)
			}
			st := scopedUserStorage(t, root, users.Permissions{Admin: true}, key)
			if err := st.Settings.Save(&settings.Settings{Key: key, Signup: true, CreateUserDir: false, Defaults: settings.UserDefaults{Scope: defaultScope, Perm: users.Permissions{Admin: true, Execute: true, Download: true, Create: true, Modify: true, Delete: true}}, MinimumPasswordLength: 1}); err != nil {
				t.Fatal(err)
			}
			signup := func(name string) int {
				req := httptest.NewRequest("POST", "/signup", strings.NewReader(`{"username":"`+name+`","password":"NewScopePassword54321!"}`))
				rec := httptest.NewRecorder()
				handle(signupHandler, "", st, &settings.Server{Root: root}).ServeHTTP(rec, req)
				return rec.Code
			}
			if status := signup("new-reader"); status != 200 {
				t.Fatalf("signup: %d", status)
			}
			// Read the persistent store directly instead of the fixture's custom FS.
			user, err := st.Users.GetByScope("/users/new-reader")
			if err != nil {
				t.Fatal(err)
			}
			if user.Perm.Admin || user.Perm.Execute {
				t.Fatal("privileges inherited")
			}
			if err := user.Clean(root, false); err != nil {
				t.Fatal(err)
			}
			if _, err := user.Fs.Open("/secret.txt"); !os.IsNotExist(err) {
				t.Fatalf("registrant can access root secret: %v", err)
			}
			if err := os.MkdirAll(filepath.Join(root, "users", "existing"), 0750); err != nil {
				t.Fatal(err)
			}
			if status := signup("existing"); status != 409 {
				t.Fatalf("pre-existing directory claimed: %d", status)
			}
		})
	}
}
