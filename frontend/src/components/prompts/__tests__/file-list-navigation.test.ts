import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import FileList from "../FileList.vue";
import { files } from "@/api";

vi.mock("@/api", () => ({ files: { fetch: vi.fn() } }));
vi.mock("@/stores/file", () => ({ useFileStore: vi.fn() }));
vi.mock("@/stores/layout", () => ({ useLayoutStore: vi.fn() }));
vi.mock("@/api/utils.js", () => ({ StatusError: class extends Error {} }));

const component = FileList as any;
const child = { url: "/files/Folder/", items: null };

function context() {
  const ctx: any = {
    ...component.data(),
    current: "/files/",
    exclude: [],
    $emit: vi.fn(),
    $showError: vi.fn(),
  };
  for (const method of ["next", "fillOptions", "abortOngoingNext"]) {
    ctx[method] = component.methods[method].bind(ctx);
  }
  return ctx;
}

function click(url = child.url, detail = 1) {
  return { currentTarget: { dataset: { url } }, detail };
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.stubGlobal("window", { location: { pathname: "/files/" } });
});
afterEach(() => vi.unstubAllGlobals());

describe("copy and move destination navigation", () => {
  it("opens a folder on one click and uses its loaded path as the destination", async () => {
    vi.mocked(files.fetch).mockResolvedValue(child as any);
    const ctx = context();
    await ctx.next(click());
    expect(files.fetch).toHaveBeenCalledExactlyOnceWith(
      child.url,
      expect.any(AbortSignal)
    );
    expect(ctx.current).toBe(child.url);
    expect(ctx.$emit).toHaveBeenCalledWith("update:selected", child.url);
    expect(ctx.$emit).toHaveBeenCalledWith("update:navigating", true);
    expect(ctx.$emit).toHaveBeenLastCalledWith("update:navigating", false);
  });

  it("ignores concurrent activation and the second click after rows are replaced", async () => {
    let resolve!: (value: any) => void;
    vi.mocked(files.fetch).mockImplementation(
      () => new Promise((done) => (resolve = done))
    );
    const ctx = context();
    const pending = ctx.next(click());
    expect(ctx.navigating).toBe(true);
    await ctx.next(click());
    expect(files.fetch).toHaveBeenCalledOnce();
    resolve(child);
    await pending;
    await ctx.next(click("/files/Folder/Nested/", 2));
    expect(files.fetch).toHaveBeenCalledOnce();
    expect(ctx.current).toBe(child.url);
    expect(ctx.navigating).toBe(false);
  });

  it("accepts keyboard activation and allows returning to the parent", async () => {
    const ctx = context();
    ctx.fillOptions(child);
    expect(ctx.items).toEqual([{ name: "..", url: "/files/" }]);
    vi.mocked(files.fetch).mockResolvedValue({
      url: "/files/",
      items: null,
    } as any);
    await ctx.next(click("/files/", 0));
    expect(ctx.current).toBe("/files/");
    expect(ctx.$emit).toHaveBeenCalledWith("update:selected", "/files/");
  });

  it("lists only allowed directories, preserving move exclusions", () => {
    const ctx = context();
    ctx.exclude = ["/files/Selected/"];
    ctx.fillOptions({
      url: "/files/",
      items: [
        { name: "Selected", url: "/files/Selected/", isDir: true },
        { name: "Folder", url: child.url, isDir: true },
        { name: "notes.md", url: "/files/notes.md", isDir: false },
      ],
    });
    expect(ctx.items).toEqual([{ name: "Folder", url: child.url }]);
  });

  it("keeps the previous destination on failure and enables retry", async () => {
    const error = new Error("unavailable");
    vi.mocked(files.fetch)
      .mockRejectedValueOnce(error)
      .mockResolvedValueOnce(child as any);
    const ctx = context();
    await ctx.next(click());
    expect(ctx.current).toBe("/files/");
    expect(ctx.navigating).toBe(false);
    expect(ctx.$showError).toHaveBeenCalledWith(error);
    expect(ctx.$emit).not.toHaveBeenCalledWith("update:selected", child.url);
    await ctx.next(click());
    expect(ctx.current).toBe(child.url);
  });

  it("aborts on closing and ignores results arriving after unmount", async () => {
    let resolve!: (value: any) => void;
    vi.mocked(files.fetch).mockImplementation(
      () => new Promise((done) => (resolve = done))
    );
    const ctx = context();
    const pending = ctx.next(click());
    const signal = ctx.nextAbortController.signal;
    component.unmounted.call(ctx);
    expect(signal.aborted).toBe(true);
    resolve(child);
    await pending;
    expect(ctx.current).toBe("/files/");
    expect(ctx.$emit).not.toHaveBeenCalledWith("update:selected", child.url);
    expect(ctx.$showError).not.toHaveBeenCalled();
  });
});
