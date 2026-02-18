import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwind from '@tailwindcss/vite';
import path from 'path';

export default defineConfig({
  root: path.resolve(__dirname, 'epn-killer-web'),
  publicDir: path.resolve(__dirname, 'epn-killer-web/public'),
  plugins: [react(), tailwind()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'epn-killer-web/src'),
    },
  },
  server: {
    port: 3000,
  },
  build: {
    outDir: path.resolve(__dirname, 'dist'),
    emptyOutDir: true,
  },
});
