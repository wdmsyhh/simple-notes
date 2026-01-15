import react from "@vitejs/plugin-react";
import { resolve } from "path";
import { defineConfig } from "vite";

let devProxyServer = "http://localhost:8080";
if (process.env.DEV_PROXY_SERVER && process.env.DEV_PROXY_SERVER.length > 0) {
  console.log("Use devProxyServer from environment: ", process.env.DEV_PROXY_SERVER);
  devProxyServer = process.env.DEV_PROXY_SERVER;
}

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: "0.0.0.0",
    port: 3000,
    proxy: {
      "^/api": {
        target: devProxyServer,
        xfwd: true,
        timeout: 60000, // 60 seconds timeout for large file uploads
        configure: (proxy, _options) => {
          proxy.on('error', (err, _req, res) => {
            console.log('proxy error', err);
          });
        },
      },
      "^/api.v1": {
        target: devProxyServer,
        xfwd: true,
        timeout: 60000, // 60 seconds timeout for large file uploads
        configure: (proxy, _options) => {
          proxy.on('error', (err, _req, res) => {
            console.log('proxy error', err);
          });
        },
      },
    },
  },
  resolve: {
    alias: {
      "@/": `${resolve(__dirname, "src")}/`,
    },
  },
});
