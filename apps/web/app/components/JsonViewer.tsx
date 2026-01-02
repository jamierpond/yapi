"use client";

import { useEffect, useRef } from "react";
import type * as Monaco from "monaco-editor";
import { monaco } from "../lib/monaco";

interface JsonViewerProps {
  value: string;
}

export default function JsonViewer({ value }: JsonViewerProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const editorRef = useRef<Monaco.editor.IStandaloneCodeEditor | null>(null);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    // If React re-runs the effect, do not re-create the editor
    if (editorRef.current) return;

    // Detect content type
    let language = "plaintext";
    try {
      JSON.parse(value);
      language = "json";
    } catch {
      // Keep as plaintext for non-JSON
    }

    // Create a model with detected language
    const model = monaco.editor.createModel(
      value,
      language
    );

    // Create the editor instance with all language features disabled
    editorRef.current = monaco.editor.create(container, {
      model,
      automaticLayout: true,
      minimap: { enabled: false },
      theme: "vs-dark",
      fontSize: 14,
      fontFamily: "var(--font-jetbrains-mono)",
      lineNumbers: "on",
      scrollBeyondLastLine: false,
      wordWrap: "on",
      padding: { top: 16, bottom: 16 },
      renderLineHighlight: "none",
      cursorBlinking: "solid",
      readOnly: true,
      domReadOnly: true,
      contextmenu: false,
      scrollbar: {
        vertical: "visible",
        horizontal: "visible",
      },
      // Disable all language features that require workers
      quickSuggestions: false,
      parameterHints: { enabled: false },
      suggestOnTriggerCharacters: false,
      acceptSuggestionOnEnter: "off",
      tabCompletion: "off",
      wordBasedSuggestions: "off",
    });

    // Cleanup on unmount
    return () => {
      editorRef.current?.dispose();
      editorRef.current = null;
      model.dispose();
    };
  }, []);

  // Update editor content and language when value prop changes
  useEffect(() => {
    if (editorRef.current) {
      const currentValue = editorRef.current.getValue();
      if (currentValue !== value) {
        // Detect new language
        let language = "plaintext";
        try {
          JSON.parse(value);
          language = "json";
        } catch {
          // Keep as plaintext
        }

        // Update model language
        const model = editorRef.current.getModel();
        if (model) {
          monaco.editor.setModelLanguage(model, language);
        }

        editorRef.current.setValue(value);
      }
    }
  }, [value]);

  return <div ref={containerRef} className="h-full w-full" />;
}
