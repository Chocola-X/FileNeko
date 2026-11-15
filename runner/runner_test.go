package runner

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// GHSA-jvpw-637p-h3pw: never expand attacker-controlled values into shell
// source. The shell reads environment variables as data at execution time.
func TestHookVariablesCannotInjectShellCode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell regression")
	}
	for _, shell := range [][]string{nil, {"/bin/sh", "-c"}} {
		for _, variable := range []string{"FILE", "USERNAME"} {
			t.Run(variable+string(rune('a'+len(shell))), func(t *testing.T) {
				dir := t.TempDir()
				marker := filepath.Join(dir, "injected")
				output := filepath.Join(dir, "output")
				value := "name\"; touch '" + marker + "'; #"
				raw := "printf '%s' \"$" + variable + "\" > \"$DESTINATION\""
				if len(shell) == 0 {
					raw = "/bin/sh -c '" + "printf \"%s\" \"$" + variable + "\" > \"$DESTINATION\"'"
				}
				r := &Runner{Enabled: true, Settings: &settings.Settings{Shell: shell}}
				if err := r.exec(raw, "after_upload", value, output, &users.User{Username: value}); err != nil {
					t.Fatal(err)
				}
				if _, err := os.Stat(marker); !os.IsNotExist(err) {
					t.Fatal("hook executed injected code")
				}
				got, err := os.ReadFile(output)
				if err != nil || string(got) != value {
					t.Fatalf("value not preserved: %q %v", got, err)
				}
			})
		}
	}
}
