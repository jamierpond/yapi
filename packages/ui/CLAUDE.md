# @yapi/ui Package

## Critical: Shared Components

The OutputPanel and JsonViewer in this package are used by BOTH:
- Web playground (`apps/web`)
- VS Code extension webview (`apps/vscode-webview`)

**Any changes here affect both.** Keep them identical - that's the whole point.

## Styling

- Font: JetBrains Mono, 14px
- Use CSS variables from `@yapi/styles/tokens.css`:
  - `--color-yapi-bg-editor` for editor background (#1e1e1e)
  - `--color-yapi-fg-linenumber` for line numbers (#858585)
