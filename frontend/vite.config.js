import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],

  // Configuração de testes com Vitest
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.js',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'src/test/',
        '**/*.test.{js,jsx}',
        '**/*.config.js',
        '**/main.jsx',
      ],
      thresholds: {
        lines: 70,
        functions: 70,
        branches: 70,
        statements: 70,
      },
    },
  },
  
  // Configurar proxy para desenvolvimento
  server: {
    port: 3000,
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8000',
        changeOrigin: true,
        secure: false,
      },
    },
  },

  // Otimizações de build para produção
  build: {
    // Output directory
    outDir: 'dist',
    
    // Generate sourcemaps apenas em desenvolvimento
    sourcemap: false,
    
    // Minificação
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true, // Remove console.logs em produção
        drop_debugger: true,
      },
    },
    
    // Tamanho máximo de chunk warning (500kb)
    chunkSizeWarningLimit: 500,
    
    // Otimizar divisão de código (code splitting)
    rollupOptions: {
      output: {
        // Separar vendor libs em chunks
        manualChunks: {
          // Core React libraries
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          // HTTP client
          'http-vendor': ['axios'],
          // UI utilities
          'ui-vendor': ['react-hot-toast'],
        },
        
        // Nomeação de chunks para melhor cache
        chunkFileNames: 'assets/js/[name]-[hash].js',
        entryFileNames: 'assets/js/[name]-[hash].js',
        assetFileNames: ({ name }) => {
          if (/\.(gif|jpe?g|png|svg|webp)$/.test(name ?? '')) {
            return 'assets/images/[name]-[hash][extname]';
          }
          if (/\.css$/.test(name ?? '')) {
            return 'assets/css/[name]-[hash][extname]';
          }
          if (/\.(woff|woff2|eot|ttf|otf)$/.test(name ?? '')) {
            return 'assets/fonts/[name]-[hash][extname]';
          }
          return 'assets/[name]-[hash][extname]';
        },
      },
    },
    
    // Reportar tamanho de chunks comprimidos
    reportCompressedSize: true,
    
    // Tamanho máximo para inlining de assets (4kb)
    assetsInlineLimit: 4096,
  },
  
  // Otimizações de preview
  preview: {
    port: 3000,
    host: true,
  },
  
  // Otimização de dependências
  optimizeDeps: {
    include: ['react', 'react-dom', 'react-router-dom', 'axios'],
  },
})
