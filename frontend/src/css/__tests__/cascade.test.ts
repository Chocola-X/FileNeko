import { describe, expect, it } from "vitest";
import { readFileSync, readdirSync } from "node:fs";
import { resolve, join } from "node:path";
import postcss from "postcss";
import { parse } from "vue/compiler-sfc";

const source = resolve(__dirname, "../..");

function files(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name);
    return entry.isDirectory() ? files(path) : [path];
  });
}

describe("customizable application styles", () => {
  it("keeps owned CSS and component styles free of forced priority", () => {
    const forced: string[] = [];
    for (const file of files(source).filter((path) =>
      /\.(css|vue)$/.test(path)
    )) {
      const text = readFileSync(file, "utf8");
      const styles = file.endsWith(".css")
        ? [text]
        : file.endsWith(".vue")
          ? parse(text).descriptor.styles.map((style) => style.content)
          : [];
      for (const style of styles) {
        postcss.parse(style, { from: file }).walkDecls((declaration) => {
          if (declaration.important) {
            forced.push(`${file}:${declaration.source?.start?.line}`);
          }
        });
      }
      // Inline bindings are outside the stylesheet cascade and must not force
      // priority either. This also covers styles injected into book previews.
      if (file.endsWith(".vue") && /!important\b/.test(text)) {
        forced.push(file);
      }
    }
    expect(forced).toEqual([]);
  });

  it("declares layer order before Vite can emit component styles", () => {
    const index = readFileSync(resolve(source, "../public/index.html"), "utf8");
    const order = index.indexOf("@layer vendor, base, components;");
    expect(order).toBeGreaterThan(-1);
    expect(order).toBeLessThan(index.indexOf('<script type="module"'));
    expect(index.indexOf('/custom.css"')).toBeGreaterThan(
      index.indexOf('<script type="module"')
    );
  });
});
