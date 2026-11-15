import { beforeEach, describe, expect, it, vi } from "vitest";
import Info from "../Info.vue";
import { files as api } from "@/api";

vi.mock("@/api", () => ({
  files: { scanDirectory: vi.fn(), checksum: vi.fn() },
}));

const prompt = Info as any;
const folder = {
  url: "/files/folder/",
  isDir: true,
  isSymlink: false,
  size: 4096,
};
const final = {
  size: 15,
  numFiles: 2,
  numDirs: 1,
  incomplete: false,
  done: true,
  queued: false,
};

function context(items = [folder]) {
  const ctx: any = { ...prompt.data(), infoItems: items };
  Object.defineProperty(ctx, "scanActive", {
    get: () => ["scanning", "queued"].includes(ctx.scanState),
  });
  ctx.startScan = prompt.methods.startScan.bind(ctx);
  ctx.stopScan = prompt.methods.stopScan.bind(ctx);
  return ctx;
}

beforeEach(() => vi.clearAllMocks());

describe("folder information scans", () => {
  it("starts on activation and combines cumulative progress without double counting", async () => {
    vi.mocked(api.scanDirectory).mockImplementation(
      async (_url, _signal, progress) => {
        progress({ ...final, size: 4, numFiles: 1, done: false });
        progress(final);
        return final;
      }
    );
    const ctx = context();
    await ctx.startScan();
    expect(ctx.scanStats).toEqual({
      size: 15,
      numFiles: 2,
      numDirs: 1,
      incomplete: false,
    });
    expect(ctx.scanState).toBe("complete");
    const startScan = vi.fn();
    prompt.activated.call({ startScan });
    expect(startScan).toHaveBeenCalledOnce();
  });

  it("aborts when closed or deactivated and ignores late results", async () => {
    let resolve!: (value: typeof final) => void;
    let progress!: (value: typeof final) => void;
    vi.mocked(api.scanDirectory).mockImplementation(
      (_url, _signal, callback) => {
        progress = callback;
        return new Promise((done) => {
          resolve = done;
        });
      }
    );
    const ctx = context();
    const running = ctx.startScan();
    const signal = ctx.scanController.signal;
    prompt.beforeUnmount.call(ctx);
    expect(signal.aborted).toBe(true);
    progress(final);
    resolve(final);
    await running;
    expect(ctx.scanStats.size).toBe(0);
    expect(ctx.scanState).toBe("canceled");
    const stopScan = vi.fn();
    prompt.deactivated.call({ stopScan });
    expect(stopScan).toHaveBeenCalledOnce();
  });

  it("starts fresh after reopening rather than reusing the previous result", async () => {
    vi.mocked(api.scanDirectory)
      .mockResolvedValueOnce(final)
      .mockResolvedValueOnce({ ...final, size: 30 });
    const ctx = context();
    await ctx.startScan();
    await ctx.startScan();
    expect(api.scanDirectory).toHaveBeenCalledTimes(2);
    expect(ctx.scanStats.size).toBe(30);
  });

  it("adds selected files and folders once, and excludes symbolic links", async () => {
    vi.mocked(api.scanDirectory).mockResolvedValue(final);
    const ctx = context([
      folder,
      { ...folder, url: "/files/two/" },
      { ...folder, isDir: false, size: 5 },
      { ...folder, isSymlink: true },
    ]);
    await ctx.startScan();
    expect(api.scanDirectory).toHaveBeenCalledTimes(2);
    expect(ctx.scanStats).toEqual({
      size: 35,
      numFiles: 5,
      numDirs: 4,
      incomplete: false,
    });
  });

  it("retains partial results and reports a failed stream", async () => {
    vi.mocked(api.scanDirectory).mockImplementation(
      async (_url, _signal, progress) => {
        progress({ ...final, done: false });
        throw new Error("disconnected");
      }
    );
    const ctx = context();
    await ctx.startScan();
    expect(ctx.scanState).toBe("failed");
    expect(ctx.scanStats.size).toBe(15);
    expect(ctx.scanController).toBeNull();
  });
});
