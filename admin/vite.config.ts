import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// Admin panel — local only. Talks to the admin API (default :8081).
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5174,
    proxy: {
      '/api': {
        target: process.env.ADMIN_API ?? 'http://localhost:8081',
        changeOrigin: true,
      },
    },
  },
})
