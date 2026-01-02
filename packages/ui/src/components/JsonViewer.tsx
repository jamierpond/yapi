"use client";

import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { vscDarkPlus } from "react-syntax-highlighter/dist/esm/styles/prism";
import type { ComponentType } from "react";

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

// Cast to work around React 19 type incompatibility
const Highlighter = SyntaxHighlighter as unknown as ComponentType<{
  language: string;
  style: Record<string, unknown>;
  showLineNumbers?: boolean;
  wrapLongLines?: boolean;
  customStyle?: Record<string, unknown>;
  lineNumberStyle?: Record<string, unknown>;
  codeTagProps?: Record<string, unknown>;
  children: string;
}>;

export default function JsonViewer({ value }: JsonViewerProps) {
  const language = detectLanguage(value);

  return (
    <div className="h-full w-full overflow-auto">
      <style>{`.linenumber { color: #858585 !important; }`}</style>
      <Highlighter
        language={language}
        style={vscDarkPlus}
        showLineNumbers
        wrapLongLines
        customStyle={{
          margin: 0,
          padding: "16px",
          background: "#1e1e1e",
          fontSize: "14px",
          fontFamily: "var(--font-jetbrains-mono), JetBrains Mono, ui-monospace, monospace",
          minHeight: "100%",
          whiteSpace: "pre-wrap",
          wordBreak: "break-word",
        }}
        lineNumberStyle={{
          minWidth: "3em",
          paddingRight: "1em",
          userSelect: "none",
        }}
        codeTagProps={{
          style: {
            whiteSpace: "pre-wrap",
            wordBreak: "break-word",
          },
        }}
      >
        {value}
      </Highlighter>
    </div>
  );
}
