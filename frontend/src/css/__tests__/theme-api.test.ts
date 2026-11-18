import { readFileSync, readdirSync } from "node:fs";
import { resolve } from "node:path";
import postcss from "postcss";
import { parse } from "vue/compiler-sfc";
import { describe, expect, it } from "vitest";

const frontend = resolve(__dirname, "../../..");
const read = (file: string) => readFileSync(resolve(frontend, file), "utf8");
const docs = read("../docs/theme-api.md");

// Keep the v1 entries when adding interfaces. Moving a component is fine;
// its public class must still mark the corresponding part of the template.
const v1Classes = {
  "views/Login.vue": "fn-page--login fn-login-panel fn-logo",
  "views/Files.vue": "fn-page--files",
  "views/Settings.vue": "fn-page--settings",
  "views/Share.vue": "fn-page--share fn-share-card",
  "views/Errors.vue": "fn-error",
  "views/files/Preview.vue": "fn-preview",
  "views/files/Editor.vue": "fn-editor",
  "views/settings/Global.vue": "fn-settings-card",
  "views/settings/Profile.vue": "fn-settings-card",
  "views/settings/Shares.vue": "fn-settings-card",
  "views/settings/User.vue": "fn-settings-card",
  "views/settings/Users.vue": "fn-settings-card",
  "components/BrandName.vue": "fn-wordmark",
  "components/header/HeaderBar.vue": "fn-header fn-logo fn-backdrop",
  "components/Sidebar.vue": "fn-sidebar fn-storage fn-backdrop",
  "components/files/ListingItem.vue":
    "fn-file-card fn-file-card__surface fn-file-card__suffix",
  "components/prompts/BaseModal.vue": "fn-dialog fn-backdrop",
  "components/prompts/PromptHeader.vue": "fn-dialog-title",
  "components/LoadingIndicator.vue": "fn-loader",
};

const v1Variables = [
  "font-ui font-heading font-mono font-signature",
  "page-background-color page-background-image page-background-size",
  "page-background-position page-background-repeat",
  "login-background-image login-background-size login-background-position login-background-repeat",
  "header-background sidebar-background login-panel-background",
  "file-card-background file-card-inner-background file-card-frame-color",
  "settings-card-background dialog-background loading-card-background card-background",
  "login-frame-opacity sidebar-frame-opacity settings-frame-opacity dialog-backdrop",
  "ink-navy ink-slate ink-muted",
  ...["cobalt", "pink", "cyan", "purple"].flatMap((tone) =>
    ["main", "dark", "light", "surface"].map((role) => `${tone}-${role}`)
  ),
].flatMap((group) => group.split(" "));

function staticClasses(template: string) {
  return [...template.matchAll(/\sclass="([^"]+)"/g)].flatMap((match) =>
    match[1].split(/\s+/)
  );
}

describe("FileNeko theme API v1", () => {
  it("keeps documented classes on their templates and on the boot screen", () => {
    for (const [file, hooks] of Object.entries(v1Classes)) {
      const template = parse(read(`src/${file}`)).descriptor.template?.content;
      expect(staticClasses(template ?? ""), file).toEqual(
        expect.arrayContaining(hooks.split(" "))
      );
      for (const hook of hooks.split(" ")) {
        expect(docs).toContain(`\`.${hook}\``);
      }
    }
    for (const file of ["index.html", "public/index.html"]) {
      const html = read(file);
      expect(html).toMatch(/<html[^>]*data-fn-theme-api="1"/);
      expect(html).toMatch(/<body class="fn-page-background"/);
      expect(html).toMatch(/id="loading"\s+class="fn-page-background"/);
      expect(staticClasses(html)).toContain("fn-loader");
    }
    expect(docs).toContain("`.fn-page-background`");
  });

  it("keeps v1 variables connected to application CSS", () => {
    const used = new Set<string>();
    for (const file of readdirSync(resolve(frontend, "src/css"))) {
      if (!file.endsWith(".css")) continue;
      postcss.parse(read(`src/css/${file}`)).walkDecls(({ value }) => {
        for (const match of value.matchAll(/var\(\s*(--[\w-]+)/g)) {
          used.add(match[1]);
        }
      });
    }
    for (const name of v1Variables) {
      expect(docs).toContain(`\`--${name}\``);
      expect(used.has(`--${name}`), `--${name} has no CSS consumer`).toBe(true);
    }
  });
});
