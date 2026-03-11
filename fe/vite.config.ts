import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      '/ws': {
        target: 'http://localhost:8080',
        ws: true,
      },
      '/pkg.rpc.ClaudeService': {
        target: 'http://localhost:8080',
      },
    },
  },
})
