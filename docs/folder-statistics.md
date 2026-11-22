# Folder Statistics

**English** | [简体中文](folder-statistics_CN.md)

Opening the "Info" dialog for the current directory or selected folders initiates a recursive scan calculating total file size, file count, and subdirectory count. For longer scans, the cumulative values update approximately every 500 milliseconds and display "Statistics complete" upon completion; there is no need to pre-calculate the total item count, and no indeterminate progress percentages are shown. When a multiple selection includes folders, each folder is scanned sequentially and combined with any selected regular files.

## Lifecycle and Storage

- Closing the dialog, pressing Escape, clicking the backdrop, switching dialogs, or navigating away causes the frontend to abort the request via `AbortController`; clicking the stop button also cancels the scan while temporarily retaining the current accumulated numbers.
- The backend listens to the HTTP request context, checking cancellation status before directory operations and between items to halt traversal and release open directory descriptors. In-flight low-level filesystem calls cannot be forcefully interrupted by Go context and exit once the call returns.
- Each server instance runs at most two concurrent scans; additional requests queue and receive waiting-state heartbeats. Queued requests can also be cancelled at any time.
- Results exist exclusively in the active request and dialog memory; no disk cache is created, no statistics database is updated, and no temporary files are generated. Reopening the dialog triggers a fresh scan without reusing stale data. Responses use `Cache-Control: no-store`.

## Calculation Rules

- Size represents the sum of file lengths rather than physical disk allocation; sparse files, filesystem compression, or deduplicated blocks may cause these values to differ. Hard links are counted individually for each directory entry.
- Accessible hidden files are included regardless of the "Show dotfiles" listing preference. Server and user path access rules remain in effect; subtrees excluded by rules are not scanned.
- Symbolic links are not traversed or counted (including when the scan root itself is a symlink) to avoid duplicate counting and infinite loops.
- When scanning a single folder, the root directory itself is not counted toward the subdirectory total; when multiple items are selected, each selected folder root is counted.
- If read errors occur, available cumulative numbers are retained with an indication that results are incomplete. Connection drops or premature cancellations are never marked as complete. Because directory contents may change during a scan, results do not represent an atomic filesystem snapshot.

## API and Deployment

`GET /api/directory-stats/<path>` requires the same authentication as file listings and streams `application/x-ndjson`. The path is relative to the user's scope; when deployed under a subpath, the frontend respects the configured `BaseURL`. The dedicated endpoint path avoids routing conflicts with regular folders named `stats`.

Each line represents cumulative progress, for example:

```json
{"size":1048576,"numFiles":12,"numDirs":3,"incomplete":false,"done":false,"queued":false}
```

`done: true` indicates scan completion; `incomplete: true` indicates that some items could not be read; `queued: true` indicates the request is awaiting an available scan slot. Clients should treat responses lacking a completion flag as interrupted and must cancel both the fetch request and its response stream whenever the info dialog is closed.

The endpoint reads metadata only, enumerating unranked directory entries in batches of up to 256 items. It does not perform MIME type detection, file content inspection, or preview generation, nor does it transmit the complete directory tree to the client.

Responses include `X-Accel-Buffering: no` to instruct Nginx to forward streaming progress without buffering. Reverse proxies should preserve the default `proxy_ignore_client_abort off` and must not ignore the `X-Accel-Buffering` response header; otherwise, closing the browser tab might allow scans to run to completion on the backend or prevent real-time progress updates.

## Verification Records (2026-11-15)

- `go test -race ./...` and `go vet ./...` passed; HTTP tests cross-compiled for Windows (not executed on Windows).
- Backend tests cover recursive sizing and counts, hidden files, access rules, symbolic links, batch enumeration, exclusion of file content reads, read error flags, descriptor cleanup on cancellation, real HTTP disconnections, concurrency limits, and queue cancellation.
- Frontend linting, 39 tests, TypeScript type checks, and production builds passed. Tests cover chunked stream parsing, premature stream termination, ignoring stale results after cancellation, component deactivation and unmount cancellation, multi-selection aggregation, and re-scanning upon reopening.
- Verified using an isolated database, temporary Nginx, and Chromium under the `/library` subpath for selected folders, current folders, empty directories, multi-selections, single file info, and re-counting after adding files. Scanning 100,000 empty files plus one sparse file yielded 100,001 files and 1 MiB; pressing Escape, clicking the backdrop, clicking OK, or clicking Stop aborted requests cleanly; 390px mobile dialogs exhibited no horizontal overflow. All temporary services and test files were cleaned up.
