import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

// Admin UI is mounted at /admin/ behind the ingress so it can coexist with
// the marketing landing page that will eventually live at /. All asset
// URLs and the React Router basename derive from this base.
export default defineConfig({
  base: '/admin/',
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5054,
  },
})
