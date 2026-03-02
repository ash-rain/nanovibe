import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: { '@': resolve(__dirname, 'src') }
  },
  server: {
    proxy: {
      '/api': 'http://localhost:3000',
      '/ws': { target: 'ws://localhost:3000', ws: true },
      '/auth': 'http://localhost:3000'
    }
  },
  build: {
    outDir: '../server/public',
    emptyOutDir: true
  }
})
