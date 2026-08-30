/// <reference types="vitest/config" />
import path from "path";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  build: {
    chunkSizeWarningLimit: 1000,
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-react': ['react', 'react-dom', 'react-router'],
          'vendor-wagmi': ['viem', 'wagmi', '@tanstack/react-query', 'axios'],
          'vendor-ui': ['lucide-react', 'class-variance-authority', 'clsx', 'tailwind-merge', 'zod'],
        },
      },
    },
  },
});
