package fbhttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/spf13/afero"
)

type directoryStatsUpdate struct {
	files.DirectoryStats
	Done   bool `json:"done"`
	Queued bool `json:"queued"`
}

// Each handler instance bounds disk work across all users. Waiting and active
// requests are both canceled by closing the information dialog/HTTP request.
func newDirectoryStatsHandler() handleFunc {
	slots := make(chan struct{}, 2)
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if r.Context().Err() != nil {
			return 0, nil
		}
		root := r.URL.Path
		if root == "" {
			root = "/"
		}
		// Dotfile hiding is a listing preference, not an access restriction.
		if !d.CheckRules(root) {
			return http.StatusForbidden, nil
		}
		var info os.FileInfo
		var err error
		if lstater, ok := d.user.Fs.(afero.Lstater); ok {
			info, _, err = lstater.LstatIfPossible(root)
		} else {
			info, err = d.user.Fs.Stat(root)
		}
		if err != nil {
			return errToStatus(err), err
		}
		if !info.IsDir() || files.IsSymlink(info.Mode()) {
			return http.StatusBadRequest, fmt.Errorf("expected a directory, not a file or symbolic link")
		}
		if _, ok := w.(http.Flusher); !ok {
			return http.StatusInternalServerError, fmt.Errorf("streaming is unavailable")
		}
		w.Header().Set("Content-Type", "application/x-ndjson; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Accel-Buffering", "no")
		encoder := json.NewEncoder(w)
		controller := http.NewResponseController(w)
		write := func(update directoryStatsUpdate) error {
			if err := r.Context().Err(); err != nil {
				return err
			}
			if err := encoder.Encode(update); err != nil {
				return err
			}
			return controller.Flush()
		}

		select {
		case slots <- struct{}{}:
		default:
			if err := write(directoryStatsUpdate{Queued: true}); err != nil {
				return 0, nil
			}
			heartbeat := time.NewTicker(5 * time.Second)
			defer heartbeat.Stop()
		waiting:
			for {
				select {
				case slots <- struct{}{}:
					break waiting
				case <-r.Context().Done():
					return 0, nil
				case <-heartbeat.C:
					if err := write(directoryStatsUpdate{Queued: true}); err != nil {
						return 0, nil
					}
				}
			}
		}
		defer func() { <-slots }()
		if err := write(directoryStatsUpdate{}); err != nil {
			return 0, nil
		}
		lastUpdate := time.Now()
		stats, err := files.ScanDirectory(r.Context(), d.user.Fs, root, d.CheckRules, func(stats files.DirectoryStats) error {
			if time.Since(lastUpdate) < 500*time.Millisecond {
				return nil
			}
			lastUpdate = time.Now()
			return write(directoryStatsUpdate{DirectoryStats: stats})
		})
		if err != nil || r.Context().Err() != nil {
			// A disconnected client or failed write must stop work, not start a
			// second HTTP response. Unexpected scan failures end the stream without
			// a done marker, which the client reports as an incomplete request.
			return 0, nil
		}
		_ = write(directoryStatsUpdate{DirectoryStats: stats, Done: true})
		return 0, nil
	})
}
