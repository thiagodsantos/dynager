import { defineConfig, type UserConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { fileURLToPath, URL } from 'node:url'

const apiTarget = process.env.DYNAGER_API_TARGET ?? 'http://localhost:8080'

const config = {
  plugins: [react()],
  publicDir: fileURLToPath(new URL('../assets', import.meta.url)),
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  test: {
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    css: true,
  },
  server: {
    proxy: {
      '/api': { target: apiTarget, changeOrigin: true },
    },
  },
} as unknown as UserConfig & { test: { environment: 'jsdom'; setupFiles: string; css: boolean } }

export default defineConfig(config)
