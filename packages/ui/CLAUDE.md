# @yapi/ui Package

## Critical: Shared Components

The OutputPanel and JsonViewer in this package are used by BOTH:
- Web playground (`apps/web`)
- VS Code extension webview (`apps/vscode-webview`)

**Any changes here affect both.** Keep them identical - that's the whole point.

## Styling

- Font: JetBrains Mono, 14px
- Background: #1e1e1e
- Line numbers: #858585 (grey) - set via `.linenumber` CSS override in both apps
