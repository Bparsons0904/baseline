import { defineConfig } from "vite";
import solidPlugin from "vite-plugin-solid";
import path from "path";

export default defineConfig({
  envDir: "../",
  plugins: [solidPlugin()],
  server: {
    port: 3010,
    host: "0.0.0.0",
    hmr: {
      port: 3010,
    },
    watch: {
      usePolling: true,
    },
  },
  resolve: {
    alias: {
      "@styles": path.resolve(__dirname, "./src/styles"),
      "@components": path.resolve(__dirname, "./src/components"),
      "@layout": path.resolve(__dirname, "./src/components/layout/"),
      "@pages": path.resolve(__dirname, "./src/pages/"),
      "@hooks": path.resolve(__dirname, "./src/hooks/"),
      "@services": path.resolve(__dirname, "./src/services/"),
      "@context": path.resolve(__dirname, "./src/context/"),
    },
  },
  build: {
    target: "esnext",
  },
  css: {
    preprocessorOptions: {
      scss: {
        includePaths: [path.resolve(__dirname, "src/styles")],
        additionalData: `
        @use "@styles/variables" as *;
        @use "@styles/mixins" as *;
        @use "@styles/colors" as *;
      `,
      },
    },
  },
});
