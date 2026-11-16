import path from "node:path";
import { readFileSync } from "node:fs";
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import VueI18nPlugin from "@intlify/unplugin-vue-i18n/vite";
import legacy from "@vitejs/plugin-legacy";
import { compression } from "vite-plugin-compression2";

const plugins = [
  {
    name: "fileneko-loading-screen",
    transformIndexHtml(html: string) {
      // The first loading frame must work before the app's JS, CSS or fonts arrive.
      return html
        .replace(
          "<!-- fileneko-loading-style -->",
          () =>
            `<style>${readFileSync(path.resolve(__dirname, "src/css/loading.css"), "utf8")}</style>`
        )
        .replace("<!-- fileneko-loading-artwork -->", () =>
          readFileSync(
            path.resolve(__dirname, "src/assets/loading-book.svg"),
            "utf8"
          )
        );
    },
  },
  vue(),
  VueI18nPlugin({
    include: [path.resolve(__dirname, "./src/i18n/**/*.json")],
  }),
  legacy({
    // defaults already drop IE support
    targets: ["defaults"],
  }),
  // The Go static handler serves .js.gz directly or decompresses it for clients
  // without gzip support. Embed one copy instead of raw JS + gzip + Brotli.
  compression({
    include: /\.js$/,
    algorithms: ["gzip"],
    deleteOriginalAssets: true,
    // The handler requires a gzip file even for tiny modules.
    skipIfLargerOrEqual: false,
  }),
];

const resolve = {
  alias: {
    // vue: "@vue/compat",
    "@/": `${path.resolve(__dirname, "src")}/`,
  },
};

// https://vitejs.dev/config/
export default defineConfig(({ command }) => {
  if (command === "serve") {
    return {
      plugins,
      resolve,
      server: {
        proxy: {
          "/api/command": {
            target: "ws://127.0.0.1:8080",
            ws: true,
          },
          "/api": "http://127.0.0.1:8080",
        },
      },
    };
  } else {
    // command === 'build'
    return {
      plugins,
      resolve,
      base: "",
      build: {
        rollupOptions: {
          input: {
            index: path.resolve(__dirname, "./public/index.html"),
          },
          output: {
            manualChunks: (id) => {
              // bundle dayjs files in a single chunk
              // this avoids having small files for each locale
              if (id.includes("dayjs/")) {
                return "dayjs";
                // bundle i18n in a separate chunk
              } else if (id.includes("i18n/")) {
                return "i18n";
              }
            },
          },
        },
      },
      experimental: {
        renderBuiltUrl(filename, { hostType }) {
          if (hostType === "js") {
            return { runtime: `window.__prependStaticUrl("${filename}")` };
          } else if (hostType === "html") {
            return `[{[ .StaticURL ]}]/${filename}`;
          } else {
            return { relative: true };
          }
        },
      },
    };
  }
});
