import type { Config } from "tailwindcss";

export default {
  content: {
    files: [
      "./index.html",
      "./src/**/*.{ts,tsx,js,jsx}",
      // Scan the workspace UI package source files
      "../../packages/ui/src/**/*.{ts,tsx,js,jsx}",
    ],
  },
  safelist: [
    // Force generation of yapi color utilities
    {
      pattern: /^(bg|text|border|from|to|via|ring|shadow|divide|decoration|outline|fill|stroke|caret)-(yapi-(bg|bg-elevated|fg|fg-muted|fg-subtle|border|accent|success|warning|error))/,
      variants: ['hover', 'focus', 'active', 'disabled'],
    },
  ],
  theme: {
    extend: {
      colors: {
        'yapi-bg': 'var(--color-yapi-bg)',
        'yapi-bg-elevated': 'var(--color-yapi-bg-elevated)',
        'yapi-fg': 'var(--color-yapi-fg)',
        'yapi-fg-muted': 'var(--color-yapi-fg-muted)',
        'yapi-fg-subtle': 'var(--color-yapi-fg-subtle)',
        'yapi-border': 'var(--color-yapi-border)',
        'yapi-accent': 'var(--color-yapi-accent)',
        'yapi-success': 'var(--color-yapi-success)',
        'yapi-warning': 'var(--color-yapi-warning)',
        'yapi-error': 'var(--color-yapi-error)',
      },
    },
  },
} satisfies Config;
