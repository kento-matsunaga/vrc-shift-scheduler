import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
  ],
  server: {
    proxy: {
      '/api': {
        // Docker Compose: backendサービスへプロキシ
        // ローカル開発: BACKEND_URL=http://localhost:8080 を設定
        target: process.env.BACKEND_URL || 'http://backend:8080',
        changeOrigin: true,
      },
    },
  },
})
