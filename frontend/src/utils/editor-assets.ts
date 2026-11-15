import ace from "ace-builds";

// Keep modes, workers, themes and snippets lazy at runtime, but include their
// URLs in the build so Vite applies the configured File Browser static prefix.
const assets = import.meta.glob<string>(
  "/node_modules/ace-builds/src-min-noconflict/**/*.js",
  { eager: true, query: "?url&no-inline", import: "default" }
);

export function setupLocalEditorAssets() {
  for (const [path, url] of Object.entries(assets)) {
    const file = path.split("/src-min-noconflict/")[1].replace(/\.js$/, "");
    let moduleId: string;
    if (file.startsWith("snippets/")) {
      moduleId = `ace/${file}`;
    } else {
      const match = /^(mode|theme|ext|keybinding|worker)-(.+)$/.exec(file);
      if (!match) continue;
      const [, kind, name] = match;
      moduleId =
        kind === "worker"
          ? `ace/mode/${name}_worker`
          : `ace/${kind === "keybinding" ? "keyboard" : kind}/${name}`;
    }
    ace.config.setModuleUrl(moduleId, url);
  }
}
