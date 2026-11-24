package fbhttp

import (
	"bytes"
	"context"
	"image/jpeg"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/spf13/afero"
)

type thumbnailFFmpeg struct {
	path   string
	seekFD bool
}

var videoThumbnailFFmpeg = sync.OnceValue(func() thumbnailFFmpeg {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return thumbnailFFmpeg{}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	protocols, err := exec.CommandContext(ctx, path, "-hide_banner", "-protocols").Output()
	if err != nil {
		return thumbnailFFmpeg{}
	}
	input, _, _ := strings.Cut(string(protocols), "Output:")
	return thumbnailFFmpeg{path: path, seekFD: slices.Contains(strings.Fields(input), "fd")}
})

var videoThumbnailSlots = make(chan struct{}, 2)

func handleVideoThumbnail(w http.ResponseWriter, r *http.Request, cache FileCache, file *files.FileInfo) (int, error) {
	ffmpeg := videoThumbnailFFmpeg()
	if ffmpeg.path == "" || !file.Mode.IsRegular() {
		return http.StatusNoContent, nil
	}
	ctx := r.Context()
	key := previewCacheKey(file, PreviewSizeThumb)
	if cache != nil {
		if poster, ok, _ := cache.Load(ctx, key); ok {
			return serveVideoThumbnail(w, r, file, poster)
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	select {
	case videoThumbnailSlots <- struct{}{}:
		defer func() { <-videoThumbnailSlots }()
	case <-ctx.Done():
		return http.StatusNoContent, nil
	}

	poster, err := ffmpeg.frame(ctx, file)
	if err != nil || len(poster) == 0 {
		return http.StatusNoContent, nil
	}
	if cache != nil {
		_ = cache.Store(ctx, key, poster)
	}
	return serveVideoThumbnail(w, r, file, poster)
}

func (ffmpeg thumbnailFFmpeg) frame(ctx context.Context, file *files.FileInfo) ([]byte, error) {
	opened, err := file.Fs.Open(file.Path)
	if err != nil {
		return nil, err
	}
	defer opened.Close()

	// Give FFmpeg the already scoped file, never a user-provided path or URL.
	// A seekable stdin supports MP4 files whose metadata sits after the video.
	input, protocol := "pipe:0", "pipe"
	stdin := opened
	for {
		wrapped, ok := stdin.(*afero.BasePathFile)
		if !ok {
			break
		}
		stdin = wrapped.File
	}
	if _, native := stdin.(*os.File); native && ffmpeg.seekFD {
		input, protocol = "fd:", "fd"
	}
	cmd := exec.CommandContext(ctx, ffmpeg.path,
		"-hide_banner", "-loglevel", "error", "-nostdin",
		"-max_alloc", "67108864", "-filter_threads", "1", "-threads", "1",
		"-protocol_whitelist", protocol, "-format_whitelist", "mov,matroska,webm,avi,mpegts,mpeg,asf,flv,ogg",
		"-i", input, "-map", "0:V:0", "-frames:v", "1",
		"-vf", "scale=256:256:force_original_aspect_ratio=decrease",
		"-threads", "1", "-c:v", "mjpeg", "-f", "image2pipe", "pipe:1")
	cmd.Stdin = stdin
	cmd.WaitDelay = time.Second
	poster, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	if _, err := jpeg.DecodeConfig(bytes.NewReader(poster)); err != nil {
		return nil, err
	}
	return poster, nil
}

func serveVideoThumbnail(w http.ResponseWriter, r *http.Request, file *files.FileInfo, poster []byte) (int, error) {
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "private")
	http.ServeContent(w, r, "thumbnail.jpg", file.ModTime, bytes.NewReader(poster))
	return 0, nil
}
