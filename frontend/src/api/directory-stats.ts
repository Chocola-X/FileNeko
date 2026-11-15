import { fetchURL, removePrefix } from "./utils";

export interface DirectoryStats {
  size: number;
  numFiles: number;
  numDirs: number;
  incomplete: boolean;
  done: boolean;
  queued: boolean;
}

// Fetch streaming cumulative totals rather than transferring every file name.
// Aborting this request also cancels its response body and the server scan.
export async function scanDirectory(
  url: string,
  signal: AbortSignal,
  onProgress: (stats: DirectoryStats) => void
): Promise<DirectoryStats> {
  const response = await fetchURL(`/api/directory-stats${removePrefix(url)}`, {
    signal,
    headers: { Accept: "application/x-ndjson" },
  });
  const reader = response.body?.getReader();
  if (!reader) throw new Error("Directory statistics stream is unavailable");
  const decoder = new TextDecoder();
  let pending = "";
  try {
    while (true) {
      if (signal.aborted) throw new DOMException("Scan canceled", "AbortError");
      const { value, done } = await reader.read();
      pending += decoder.decode(value, { stream: !done });
      const lines = pending.split("\n");
      pending = lines.pop()!;
      if (done && pending.trim()) lines.push(pending);
      for (const line of lines) {
        if (!line.trim()) continue;
        const stats = JSON.parse(line) as DirectoryStats;
        if (
          ![stats.size, stats.numFiles, stats.numDirs].every(
            (n) => Number.isFinite(n) && n >= 0
          ) ||
          typeof stats.done !== "boolean" ||
          typeof stats.incomplete !== "boolean" ||
          typeof stats.queued !== "boolean"
        ) {
          throw new Error("Invalid directory statistics response");
        }
        if (signal.aborted)
          throw new DOMException("Scan canceled", "AbortError");
        onProgress(stats);
        if (stats.done) return stats;
      }
      if (done) throw new Error("Directory statistics stream ended early");
      if (pending.length > 65536) throw new Error("Invalid statistics stream");
    }
  } finally {
    await reader.cancel().catch(() => {});
    reader.releaseLock();
  }
}
