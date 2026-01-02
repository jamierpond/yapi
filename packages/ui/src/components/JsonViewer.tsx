"use client";

/**
 * JsonViewer - Syntax-highlighted code viewer
 *
 * Uses the light build of react-syntax-highlighter to minimize bundle size.
 * Only json language is registered since that's all we need for API responses.
 *
 * NOTE: The React 19 type issue with react-syntax-highlighter is a known issue.
 * See: https://github.com/react-syntax-highlighter/react-syntax-highlighter/issues/539
 * Once the library updates its types for React 19, the type assertion can be removed.
 */

import { PrismLight as SyntaxHighlighter } from "react-syntax-highlighter";
import json from "react-syntax-highlighter/dist/esm/languages/prism/json";
import { vscDarkPlus } from "react-syntax-highlighter/dist/esm/styles/prism";
import type { CSSProperties, FC, ReactNode } from "react";

// Register only the languages we need (reduces bundle size significantly)
SyntaxHighlighter.registerLanguage("json", json);

interface JsonViewerProps {
  value: string;
}

// Detect if content is valid JSON
function detectLanguage(value: string): string {
  try {
    JSON.parse(value);
    return "json";
  } catch {
    return "text";
  }
}

// Type definition for SyntaxHighlighter props
// This is needed because react-syntax-highlighter types are not compatible with React 19
interface HighlighterProps {
  language: string;
  style: Record<string, CSSProperties>;
  showLineNumbers?: boolean;
  wrapLongLines?: boolean;
  customStyle?: CSSProperties;
  lineNumberStyle?: CSSProperties;
  codeTagProps?: { style?: CSSProperties };
  children: ReactNode;
}

// Type assertion for React 19 compatibility
// TODO: Remove this once react-syntax-highlighter updates its types for React 19
const Highlighter = SyntaxHighlighter as unknown as FC<HighlighterProps>;

// Strip background colors from all theme tokens to prevent banding
const customTheme: Record<string, CSSProperties> = {};
for (const [key, value] of Object.entries(vscDarkPlus)) {
  if (typeof value === "object" && value !== null) {
    const { background, backgroundColor, ...rest } = value as Record<string, unknown>;
    customTheme[key] = rest as CSSProperties;
  }
}
customTheme['pre[class*="language-"]'] = {
  background: "var(--color-yapi-bg-editor)",
  margin: 0,
};
customTheme['code[class*="language-"]'] = {
  background: "transparent",
};

export default function JsonViewer({ value }: JsonViewerProps) {
  const language = detectLanguage(value);

  return (
    <div className="h-full w-full overflow-auto" style={{ background: "var(--color-yapi-bg-editor)" }}>
      <Highlighter
        language={language}
        style={customTheme}
        showLineNumbers
        wrapLongLines={false}
        customStyle={{
          margin: 0,
          padding: "16px",
          background: "var(--color-yapi-bg-editor)",
          fontSize: "var(--vscode-editor-font-size, 14px)",
          fontFamily: "var(--font-jetbrains-mono), JetBrains Mono, ui-monospace, monospace",
          minHeight: "100%",
          whiteSpace: "pre",
        }}
        lineNumberStyle={{
          minWidth: "3em",
          paddingRight: "1em",
          userSelect: "none",
          color: "var(--color-yapi-fg-linenumber)",
        }}
      >
        {value}
      </Highlighter>
    </div>
  );
}
