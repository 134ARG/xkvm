import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import tailwindcss from "@tailwindcss/vite";
import tsconfigPaths from "vite-tsconfig-paths";
import basicSsl from "@vitejs/plugin-basic-ssl";
import { paraglideVitePlugin } from "@inlang/paraglide-js";

declare const process: {
  env: {
    XKVM_PROXY_URL: string;
    USE_SSL: string;
  };
};

export default defineConfig(({ mode, command }) => {
  const isCloud = mode.indexOf("cloud") !== -1;
  const onDevice = mode === "device";
  const { XKVM_PROXY_URL, USE_SSL } = process.env;
  const useSSL = USE_SSL === "true";

  const plugins = [
    tailwindcss(),
    tsconfigPaths(),
    react()
  ];

  if (useSSL) {
    plugins.push(basicSsl());
  }

  plugins.push(paraglideVitePlugin({
    project: "./localization/xKVM.UI.inlang",
    outdir: "./localization/paraglide",
    outputStructure: 'message-modules',
    cookieName: 'XKVM_LOCALE',
    strategy: ['cookie', 'baseLocale'],
  }))

  return {
    plugins,
    esbuild: {
      pure: command === "build" ? ["console.debug"]: [],
    },
    assetsInclude: ["**/*.woff2"],
    build: {
      outDir: isCloud ? "dist" : "../static",
      rollupOptions: {
        output: {
          manualChunks: (id) => {
            if (id.includes("node_modules")) {
              return "vendor";
            }
            return null;
          },
          assetFileNames: "assets/immutable/[name]-[hash][extname]",
          chunkFileNames: "assets/immutable/[name]-[hash].js",
          entryFileNames: "assets/immutable/[name]-[hash].js",
        },
      },
    },
    server: {
      host: "0.0.0.0",
      https: useSSL,
      proxy: XKVM_PROXY_URL
        ? {
          "/me": XKVM_PROXY_URL,
          "/device": XKVM_PROXY_URL,
          "/webrtc": XKVM_PROXY_URL,
          "/auth": XKVM_PROXY_URL,
          "/storage": XKVM_PROXY_URL,
          "/cloud": XKVM_PROXY_URL,
          "/developer": XKVM_PROXY_URL,
        }
        : undefined,
    },
    base: onDevice && command === "build" ? "/static" : "/",
  };
});
