package fbhttp

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
