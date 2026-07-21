import { fileURLToPath, URL } from "node:url"

import { defineConfig } from "vite"
import vue from "@vitejs/plugin-vue"
import vueDevTools from "vite-plugin-vue-devtools"
import tailwindcss from "@tailwindcss/vite"

// https://vite.dev/config/
export default defineConfig(({ command }) => ({
  base: command === "build" ? "/app" : "/",
  plugins: [vue(), vueDevTools(), tailwindcss()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url))
    }
  },
  server: {
    allowedHosts: true,
    host: true,
    port: 5173,
    proxy: {
      "/api": "http://backend:3000",
      "/upload": "http://backend:3000",
      "/set-password": "http://backend:3000"
    }
  }
}))
