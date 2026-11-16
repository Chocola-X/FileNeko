package fbhttp

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/asticode/go-astisub"

	"github.com/filebrowser/filebrowser/v2/files"
)

var srtLineBreakTag = regexp.MustCompile(`(?i)<br(?:\s+[^>]*)?\s*/?>`)

const maxSubtitleSize = 10 << 20 // Bound conversion memory, including SSA/ASS parsing.

var subtitleHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Download {
		return http.StatusAccepted, nil
	}

	file, err := files.NewFileInfo(&files.FileOptions{
		Fs:         d.user.Fs,
		Path:       r.URL.Path,
		Modify:     d.user.Perm.Modify,
		Expand:     false,
		ReadHeader: d.server.TypeDetectionByHeader,
		Checker:    d,
	})
	if err != nil {
		return errToStatus(err), err
	}

	if file.IsDir {
		return http.StatusBadRequest, nil
	}

	return subtitleFileHandler(w, r, file)
})

func subtitleFileHandler(w http.ResponseWriter, r *http.Request, file *files.FileInfo) (int, error) {
	// if its not a subtitle file, reject
	ext := strings.ToLower(filepath.Ext(file.Name))
	if ext != ".srt" && ext != ".ass" && ext != ".ssa" && ext != ".vtt" {
		return http.StatusBadRequest, nil
	}

	info, err := file.Fs.Stat(file.Path)
	if err != nil {
		return errToStatus(err), err
	}
	if !info.Mode().IsRegular() {
		return http.StatusBadRequest, nil
	}

	fd, err := file.Fs.Open(file.Path)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer fd.Close()
	info, err = fd.Stat()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if !info.Mode().IsRegular() {
		return http.StatusBadRequest, nil
	}

	var content []byte
	if ext != ".vtt" {
		if info.Size() > maxSubtitleSize {
			return http.StatusRequestEntityTooLarge, nil
		}
		// Stat alone does not protect against a file growing during the read.
		content, err = io.ReadAll(io.LimitReader(fd, maxSubtitleSize+1))
		if err != nil {
			return http.StatusInternalServerError, err
		}
		if len(content) > maxSubtitleSize {
			return http.StatusRequestEntityTooLarge, nil
		}
	}

	// load subtitle for conversion to vtt
	var sub *astisub.Subtitles
	switch ext {
	case ".srt":
		sub, err = astisub.ReadFromSRT(bytes.NewReader(normalizeSRTLineBreaks(content)))
	case ".ass", ".ssa":
		sub, err = astisub.ReadFromSSA(bytes.NewReader(content))
	}
	if err != nil {
		return http.StatusInternalServerError, err
	}

	setContentDisposition(w, r, file)
	w.Header().Add("Content-Security-Policy", `script-src 'none';`)
	w.Header().Set("Cache-Control", "private")
	// force type to text/vtt
	w.Header().Set("Content-Type", "text/vtt")

	// serve vtt file directly
	if sub == nil {
		http.ServeContent(w, r, file.Name, file.ModTime, fd)
		return 0, nil
	}

	// convert others to vtt and serve from buffer
	var buf = &bytes.Buffer{}
	err = sub.WriteToWebVTT(buf)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	http.ServeContent(w, r, file.Name, file.ModTime, bytes.NewReader(buf.Bytes()))
	return 0, nil
}

func normalizeSRTLineBreaks(content []byte) []byte {
	return srtLineBreakTag.ReplaceAll(content, []byte("\n"))
}
