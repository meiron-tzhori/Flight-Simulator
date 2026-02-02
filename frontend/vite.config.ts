import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // Proxy ALL backend API calls to Go server (localhost:8080)
      '/stream': {
        target: 'http://localhost:8080',
        ws: true,  // Enable WebSocket (SSE works over WS)
        changeOrigin: true,
      },
      '/command': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/state': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/health': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      // Optional: proxy root (if you serve API at root)
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      }
    }
  }
})
