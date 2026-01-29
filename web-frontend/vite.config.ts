import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import Sitemap from 'vite-plugin-sitemap'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    Sitemap({
      hostname: 'https://vrcshift.com',
      // Public pages only - authenticated routes are excluded
      dynamicRoutes: ['/terms', '/privacy', '/subscribe'],
      exclude: [
        '/admin/*',
        '/events/*',
        '/members',
        '/attendance/*',
        '/schedules/*',
        '/settings',
        '/p/*',
        '/invite/*',
        '/register',
        '/reset-password',
        '/forgot-password',
      ],
      changefreq: 'weekly',
      priority: 0.8,
      lastmod: new Date(),
      generateRobotsTxt: false, // Use custom robots.txt from public/
    }),
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
