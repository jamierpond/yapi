import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import tsconfigPaths from 'vite-tsconfig-paths'

export default defineConfig({
  plugins: [tsconfigPaths(), react()],
  test: {
    name: 'browser',
    environment: 'jsdom',
    include: ['__test__/browser/**/*.test.{ts,tsx}'],
    globals: true,
  },
})
