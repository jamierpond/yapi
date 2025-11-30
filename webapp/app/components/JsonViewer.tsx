"use client";

import { useEffect, useRef } from "react";
import * as monaco from "monaco-editor";

interface JsonViewerProps {
  value: string;
}

export default function JsonViewer({ value }: JsonViewerProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const editorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    // If React re-runs the effect, do not re-create the editor
    if (editorRef.current) return;

    // Configure workers once
    if (typeof window !== "undefined" && !(window as any).MonacoEnvironment) {
      (window as any).MonacoEnvironment = {
        getWorker(_id: string, label: string) {
          // Fallback to default Monaco editor worker
          return new Worker(
            new URL(
              "monaco-editor/esm/vs/editor/editor.worker",
              import.meta.url
            ),
            { type: "module" }
          );
        },
      };
    }

    // Create a JSON model
    const model = monaco.editor.createModel(
      value,
      "json",
      monaco.Uri.parse("file:///response.json")
    );

    // Create the editor instance
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
    });

    // Cleanup on unmount
    return () => {
      editorRef.current?.dispose();
      editorRef.current = null;
      model.dispose();
    };
  }, []);

  // Update editor content when value prop changes
  useEffect(() => {
    if (editorRef.current) {
      const currentValue = editorRef.current.getValue();
      if (currentValue !== value) {
        editorRef.current.setValue(value);
      }
    }
  }, [value]);

  return <div ref={containerRef} className="h-full w-full" />;
}
