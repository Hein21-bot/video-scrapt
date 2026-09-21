import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    host: true,          // bind 0.0.0.0 — reachable from phones on the same Wi-Fi
    proxy: {
      '/api': {
        target: process.env.API_PROXY ?? 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
