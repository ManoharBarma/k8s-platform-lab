import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// In dev, proxy /api to the local api-gateway so the SPA and API share an origin.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: process.env.VITE_GATEWAY_URL || 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
