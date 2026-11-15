import { beforeEach, describe, expect, it, vi } from "vitest";
import { scanDirectory, type DirectoryStats } from "../directory-stats";
import { fetchURL } from "../utils";

vi.mock("../utils", () => ({
  fetchURL: vi.fn(),
  removePrefix: (url: string) => url.replace(/^\/files/, ""),
}));

const initial: DirectoryStats = {
  size: 0,
  numFiles: 0,
  numDirs: 0,
  incomplete: false,
  queued: false,
  done: false,
};
const final = { ...initial, size: 15, numFiles: 2, numDirs: 1, done: true };

function stream(parts: string[], close = true, cancel = vi.fn()) {
  const body = new ReadableStream<Uint8Array>({
    start(controller) {
      for (const part of parts)
        controller.enqueue(new TextEncoder().encode(part));
      if (close) controller.close();
    },
    cancel,
  });
  vi.mocked(fetchURL).mockResolvedValue(new Response(body));
  return cancel;
}

beforeEach(() => vi.clearAllMocks());

describe("directory statistics stream", () => {
  it("handles split records, queue updates and the final total", async () => {
    const queued = { ...initial, queued: true };
    const partial = { ...initial, size: 5, numFiles: 1 };
    const data = [queued, partial, final]
      .map((s) => JSON.stringify(s))
      .join("\n");
    stream([data.slice(0, 7), data.slice(7, 150), data.slice(150)]);
    const controller = new AbortController();
    const progress = vi.fn();
    expect(
      await scanDirectory("/files/folder%20name/", controller.signal, progress)
    ).toEqual(final);
    expect(fetchURL).toHaveBeenCalledWith(
      "/api/directory-stats/folder%20name/",
      expect.objectContaining({ signal: controller.signal })
    );
    expect(progress.mock.calls.map(([s]) => s)).toEqual([
      queued,
      partial,
      final,
    ]);
  });

  it("rejects truncated results instead of reporting a completed scan", async () => {
    stream([JSON.stringify(initial) + "\n"]);
    await expect(
      scanDirectory("/files/", new AbortController().signal, vi.fn())
    ).rejects.toThrow("ended early");
  });

  it("cancels the reader and suppresses buffered results after abort", async () => {
    const canceled = stream(
      [JSON.stringify(initial) + "\n" + JSON.stringify(final) + "\n"],
      false
    );
    const controller = new AbortController();
    const progress = vi.fn(() => controller.abort());
    await expect(
      scanDirectory("/files/", controller.signal, progress)
    ).rejects.toMatchObject({ name: "AbortError" });
    expect(progress).toHaveBeenCalledTimes(1);
    expect(canceled).toHaveBeenCalledTimes(1);
  });

  it("rejects invalid totals", async () => {
    stream([JSON.stringify({ ...final, size: -1 }) + "\n"]);
    await expect(
      scanDirectory("/files/", new AbortController().signal, vi.fn())
    ).rejects.toThrow("Invalid directory statistics");
  });
});
