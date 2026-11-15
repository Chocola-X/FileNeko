import { describe, expect, it, vi } from "vitest";

const { setModuleUrl } = vi.hoisted(() => ({ setModuleUrl: vi.fn() }));
vi.mock("ace-builds", () => ({ default: { config: { setModuleUrl } } }));

import { setupLocalEditorAssets } from "../editor-assets";

describe("self-hosted editor resources", () => {
  it("registers executable local URLs for syntax, workers and optional modules", () => {
    setupLocalEditorAssets();
    const modules = new Map<string, string>(setModuleUrl.mock.calls);
    for (const id of [
      "ace/mode/json",
      "ace/mode/javascript",
      "ace/mode/json_worker",
      "ace/mode/base_worker",
      "ace/theme/chrome",
      "ace/snippets/javascript",
      "ace/keyboard/vim",
      "ace/ext/searchbox",
    ]) {
      expect(modules.has(id), id).toBe(true);
    }
    for (const url of modules.values()) {
      expect(url).not.toMatch(/^(?:https?:|\/\/|data:)/i);
      expect(url).toMatch(/\.js(?:\?|$)/);
    }
  });
});
