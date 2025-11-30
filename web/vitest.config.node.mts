import { defineConfig } from 'vitest/config'
import tsconfigPaths from 'vite-tsconfig-paths'

export default defineConfig({
  plugins: [tsconfigPaths()],
  test: {
    name: 'node',
    environment: 'node',
    include: ['__test__/node/**/*.test.{ts,tsx}'],
    globals: true,
  },
})
