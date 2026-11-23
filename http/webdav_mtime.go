package fbhttp

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
)

const davMtimeHeader = "X-OC-Mtime"

var davLastModified = xml.Name{Space: "DAV:", Local: "lastmodified"}

func parseDAVMtime(value string) (time.Time, error) {
	seconds, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	// Bound the year to the four digits used in WebDAV HTTP dates.
	if err != nil || seconds < -62135596800 || seconds > 253402300799 {
		return time.Time{}, fmt.Errorf("invalid WebDAV modification time %q", value)
	}
	return time.Unix(seconds, 0), nil
}

func (f *davFS) setMtime(name string, mtime time.Time) error {
	if err := f.d.user.Fs.Chtimes(name, time.Now(), mtime); err != nil {
		return err
	}
	info, err := f.d.user.Fs.Stat(name)
	if err != nil {
		return err
	}
	if !info.ModTime().Equal(mtime) {
		return fmt.Errorf("filesystem did not preserve modification time %s", mtime)
	}
	return nil
}

type davMtimeResponse struct {
	http.ResponseWriter
	accepted bool
}

func (w *davMtimeResponse) WriteHeader(status int) {
	if status >= 200 && status < 300 && w.accepted {
		w.Header().Set(davMtimeHeader, "accepted")
	}
	w.ResponseWriter.WriteHeader(status)
}

type davMtimeUploadFile struct {
	*davFile
	written func(string)
}

func (f *davMtimeUploadFile) ReadFrom(r io.Reader) (int64, error) {
	n, err := io.Copy(f.File, r)
	if err == nil {
		// x/net computes the PUT ETag from Stat after the copy, under the lock.
		f.written(f.name)
	}
	return n, err
}

type davMtimePatchFile struct {
	afero.File
	set func(time.Time) error
}

func (f *davMtimePatchFile) DeadProps() (map[xml.Name]webdav.Property, error) {
	return nil, nil
}

// The ownCloud DAV:lastmodified property lets clients restore directory
// timestamps after uploading children. DAV:getlastmodified remains read-only.
func (f *davMtimePatchFile) Patch(patches []webdav.Proppatch) ([]webdav.Propstat, error) {
	var forbidden, modified []webdav.Property
	for _, patch := range patches {
		for _, prop := range patch.Props {
			if prop.XMLName == davLastModified && !patch.Remove {
				modified = append(modified, prop)
			} else {
				forbidden = append(forbidden, prop)
			}
		}
	}
	if len(forbidden) > 0 {
		stats := []webdav.Propstat{{Status: http.StatusForbidden, Props: forbidden}}
		if len(modified) > 0 {
			stats = append(stats, webdav.Propstat{Status: webdav.StatusFailedDependency, Props: modified})
		}
		return stats, nil
	}
	if len(modified) != 1 {
		return []webdav.Propstat{{Status: http.StatusConflict, Props: modified}}, nil
	}
	mtime, err := parseDAVMtime(string(modified[0].InnerXML))
	if err != nil {
		return []webdav.Propstat{{Status: http.StatusConflict, Props: modified}}, nil
	}
	if err := f.set(mtime); err != nil {
		return nil, err
	}
	return []webdav.Propstat{{Status: http.StatusOK, Props: modified}}, nil
}
