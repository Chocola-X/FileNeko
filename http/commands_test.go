package fbhttp

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/gorilla/websocket"
)

// GHSA-39cx-23x9-5c8p: reject unauthorized clients before upgrading and bound
// complete messages (the upgrader's read buffer does not impose a size limit).
func TestCommandWebSocketLimits(t *testing.T) {
	for _, tc := range []struct {
		name               string
		enabled, permitted bool
	}{
		{"disabled", false, true}, {"denied", true, false}, {"allowed", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			key := []byte("test-signing-key")
			perm := users.Permissions{Execute: tc.permitted}
			st := scopedUserStorage(t, t.TempDir(), perm, key)
			srv := httptest.NewServer(handle(commandsHandler, "", st, &settings.Server{EnableExec: tc.enabled}))
			defer srv.Close()
			h := http.Header{"X-Auth": {signToken(t, perm, key, st)}}
			conn, resp, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/", h)
			if !tc.enabled || !tc.permitted {
				if conn != nil {
					conn.Close()
				}
				if err == nil || resp == nil || resp.StatusCode != http.StatusForbidden {
					t.Fatalf("unauthorized upgrade: response=%v err=%v", resp, err)
				}
				resp.Body.Close()
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			_ = conn.WriteMessage(websocket.TextMessage, bytes.Repeat([]byte("x"), maxCommandSize+1))
			_, _, err = conn.ReadMessage()
			if !websocket.IsCloseError(err, websocket.CloseMessageTooBig) {
				t.Fatalf("oversized message: got %v, want close 1009", err)
			}
		})
	}
}

// GHSA-8c9q-7855-wfxq: configuring a shell must not turn the unchecked tail
// of an allowed command into a second command.
func TestCommandMetacharactersAreNotShellCode(t *testing.T) {
	echo, err := exec.LookPath("echo")
	if err != nil {
		t.Skip("echo unavailable")
	}
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell regression")
	}
	root := t.TempDir()
	marker := filepath.Join(root, "injected")
	key := []byte("key")
	perm := users.Permissions{Execute: true}
	st := scopedUserStorage(t, root, perm, key)
	user, err := st.Users.Get("", false, uint(1))
	if err != nil {
		t.Fatal(err)
	}
	user.Commands = []string{echo}
	if err := st.Users.Update(user, "Commands"); err != nil {
		t.Fatal(err)
	}
	if err := st.Settings.Save(&settings.Settings{Key: key, Shell: []string{"/bin/sh", "-c"}}); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(handle(commandsHandler, "", st, &settings.Server{EnableExec: true}))
	defer srv.Close()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/", http.Header{"X-Auth": {signToken(t, perm, key, st)}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.WriteMessage(websocket.TextMessage, []byte(echo+" safe; touch '"+marker+"'")); err != nil {
		t.Fatal(err)
	}
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	_, output, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(output), "safe") {
		t.Fatalf("allowed command did not run: %q", output)
	}
	// Wait for handler completion before checking for a chained command.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("shell injection created %s", marker)
	}
}
