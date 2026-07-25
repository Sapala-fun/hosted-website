import { defineConfig } from 'astro/config';

export default defineConfig({
  base: '/',
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:3001',
        changeOrigin: true,
      },
    },
  },
});
