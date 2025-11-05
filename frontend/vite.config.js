import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  
  // Configurar proxy para desenvolvimento
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8000',
        changeOrigin: true,
        secure: false,
      },
    },
  },

  // Otimizações de build
  build: {
    // Tamanho máximo de chunk warning (500kb)
    chunkSizeWarningLimit: 500,
    
    // Otimizar divisão de código
    rollupOptions: {
      output: {
        manualChunks: {
          // Separar vendor libs em chunks
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          'utils': ['axios', 'date-fns'],
        },
      },
    },
  },
})
