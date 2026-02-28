import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    host: '0.0.0.0',
    proxy: {
      '/v1': 'http://localhost:8686',
      '/admin': 'http://localhost:8686',
      '/health': 'http://localhost:8686',
    },
    watch: {
      usePolling: true,
      interval: 500,
    },
  },
})
