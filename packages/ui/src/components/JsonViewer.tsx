"use client";

import { useEffect, useRef, useState } from "react";
import * as monaco from "monaco-editor";

// Configure Monaco workers for Vite
self.MonacoEnvironment = {
  getWorker(_: unknown, label: string) {
    if (label === "json") {
      return new Worker(
        new URL("monaco-editor/esm/vs/language/json/json.worker.js", import.meta.url),
        { type: "module" }
      );
    }
    return new Worker(
      new URL("monaco-editor/esm/vs/editor/editor.worker.js", import.meta.url),
      { type: "module" }
    );
  },
};

interface JsonViewerProps {
  value: string;
}

export default function JsonViewer({ value }: JsonViewerProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const editorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    // If React re-runs the effect, do not re-create the editor
    if (editorRef.current) return;

    // Delay initialization to let container settle
    const initTimeout = setTimeout(() => {
      // Detect content type
      let language = "plaintext";
      try {
        JSON.parse(value);
        language = "json";
      } catch {
        // Keep as plaintext for non-JSON
      }

      // Create a model with detected language
      const model = monaco.editor.createModel(value, language);

      // Create the editor instance - disable automaticLayout, we'll handle it manually
      editorRef.current = monaco.editor.create(container, {
        model,
        automaticLayout: false,
        minimap: { enabled: false },
        theme: "vs-dark",
        fontSize: 13,
        fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace",
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
          useShadows: false,
        },
        overviewRulerLanes: 0,
        hideCursorInOverviewRuler: true,
        overviewRulerBorder: false,
        // Disable all language features that require workers
        quickSuggestions: false,
        parameterHints: { enabled: false },
        suggestOnTriggerCharacters: false,
        acceptSuggestionOnEnter: "off",
        tabCompletion: "off",
        wordBasedSuggestions: "off",
      });

      // Initial layout
      editorRef.current.layout();
      setIsReady(true);
    }, 50);

    // Manual resize handling with ResizeObserver
    const resizeObserver = new ResizeObserver(() => {
      if (editorRef.current) {
        // Use requestAnimationFrame to debounce and sync with rendering
        requestAnimationFrame(() => {
          editorRef.current?.layout();
        });
      }
    });
    resizeObserver.observe(container);

    // Cleanup on unmount
    return () => {
      clearTimeout(initTimeout);
      resizeObserver.disconnect();
      editorRef.current?.dispose();
      editorRef.current = null;
    };
  }, []);

  // Update editor content and language when value prop changes
  useEffect(() => {
    if (editorRef.current && isReady) {
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
  }, [value, isReady]);

  return <div ref={containerRef} className="h-full w-full" />;
}
