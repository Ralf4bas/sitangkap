import { defineConfig } from 'astro/config';

export default defineConfig({
  server: { port: 4321 },
  vite: {
    server: {
      proxy: {
        '/api': 'http://localhost:3000',
        '/health': 'http://localhost:3000',
      },
    },
  },
});
