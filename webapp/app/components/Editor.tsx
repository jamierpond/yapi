"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import * as monaco from "monaco-editor";
import type { ValidateResponse, Diagnostic } from "../types/api-contract";

interface EditorProps {
  value: string;
  onChange: (value: string) => void;
  onRun: () => void;
}

export default function Editor({ value, onChange, onRun }: EditorProps) {
  // Ref to the DOM node Monaco will render into
  const containerRef = useRef<HTMLDivElement | null>(null);
  // Ref to keep the editor instance (avoid double init, allow dispose)
  const editorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);
  // Track validation state
  const [hasErrors, setHasErrors] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string>("");

  // Use ref to always have the latest onRun callback and validation state
  const onRunRef = useRef(onRun);
  const hasErrorsRef = useRef(hasErrors);
  useEffect(() => {
    onRunRef.current = onRun;
  }, [onRun]);
  useEffect(() => {
    hasErrorsRef.current = hasErrors;
  }, [hasErrors]);

  // Ref for debounce timer
  const validateTimerRef = useRef<NodeJS.Timeout | null>(null);

  // Convert API diagnostic to Monaco marker
  const diagnosticToMarker = useCallback((d: Diagnostic, model: monaco.editor.ITextModel): monaco.editor.IMarkerData => {
    const severity = d.severity === "error"
      ? monaco.MarkerSeverity.Error
      : d.severity === "warning"
        ? monaco.MarkerSeverity.Warning
        : monaco.MarkerSeverity.Info;

    // Handle line -1 (unknown position) by defaulting to line 1
    const line = d.line >= 0 ? d.line + 1 : 1;
    const col = d.col >= 0 ? d.col + 1 : 1;

    // Get the line content to determine end column
    const lineContent = model.getLineContent(Math.min(line, model.getLineCount()));
    const endCol = lineContent.length + 1;

    return {
      severity,
      message: d.message,
      startLineNumber: line,
      startColumn: col,
      endLineNumber: line,
      endColumn: endCol,
      source: "yapi",
    };
  }, []);

  // Validate via API and set markers
  const validateContent = useCallback(async (content: string) => {
    const model = editorRef.current?.getModel();
    if (!model) return;

    try {
      const response = await fetch("/api/validate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ yaml: content }),
      });

      const result: ValidateResponse = await response.json();

      // Convert diagnostics to Monaco markers
      const markers = result.diagnostics.map(d => diagnosticToMarker(d, model));

      // Add warnings as info markers
      result.warnings.forEach((w, i) => {
        markers.push({
          severity: monaco.MarkerSeverity.Warning,
          message: w,
          startLineNumber: 1,
          startColumn: 1,
          endLineNumber: 1,
          endColumn: model.getLineContent(1).length + 1,
          source: "yapi",
        });
      });

      // Set markers on the model
      monaco.editor.setModelMarkers(model, "yapi", markers);

      // Update error state
      const errors = result.diagnostics.filter(d => d.severity === "error");
      if (errors.length > 0) {
        setHasErrors(true);
        setErrorMessage(errors[0].message);
      } else {
        setHasErrors(false);
        setErrorMessage("");
      }
    } catch (err) {
      console.error("Validation error:", err);
    }
  }, [diagnosticToMarker]);

  // Debounced validation
  const debouncedValidate = useCallback((content: string) => {
    if (validateTimerRef.current) {
      clearTimeout(validateTimerRef.current);
    }
    validateTimerRef.current = setTimeout(() => {
      validateContent(content);
    }, 300);
  }, [validateContent]);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    // If React re-runs the effect (StrictMode), do not re-create the editor
    if (editorRef.current) return;

    // Configure workers once; Monaco uses this to spawn background workers
    if (typeof window !== "undefined" && !(window as any).MonacoEnvironment) {
      (window as any).MonacoEnvironment = {
        getWorker(_id: string, label: string) {
          if (label === "yaml") {
            return new Worker(
              new URL("monaco-yaml/yaml.worker", import.meta.url),
              { type: "module" }
            );
          }
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

    // Create a YAML model; URI is just a fake file name
    const model = monaco.editor.createModel(
      value,
      "yaml",
      monaco.Uri.parse("file:///example.yaml")
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
      renderLineHighlight: "all",
      cursorBlinking: "smooth",
      // ensure IntelliSense is on
      quickSuggestions: { other: true, comments: false, strings: true },
      suggestOnTriggerCharacters: true,
      acceptSuggestionOnEnter: "on",
      tabCompletion: "on",
    });

    // Listen to content changes and trigger validation
    const disposable = editorRef.current.onDidChangeModelContent(() => {
      const currentValue = editorRef.current?.getValue() || "";
      onChange(currentValue);
      debouncedValidate(currentValue);
    });

    // Run initial validation
    debouncedValidate(value);

    // Add keyboard shortcut for Cmd+Enter (Mac) or Ctrl+Enter (Windows/Linux)
    editorRef.current.addCommand(
      monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter,
      () => {
        const currentValue = editorRef.current?.getValue() || "";
        const model = editorRef.current?.getModel();

        if (model) {
          const markers = monaco.editor.getModelMarkers({ resource: model.uri });
          const errors = markers.filter(m => m.severity === monaco.MarkerSeverity.Error);

          if (errors.length === 0 && currentValue.trim()) {
            onRunRef.current();
          }
        }
      }
    );

    // Add keyboard shortcut for Cmd+S (Mac) or Ctrl+S (Windows/Linux)
    editorRef.current.addCommand(
      monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS,
      () => {
        const currentValue = editorRef.current?.getValue() || "";
        const model = editorRef.current?.getModel();

        if (model) {
          const markers = monaco.editor.getModelMarkers({ resource: model.uri });
          const errors = markers.filter(m => m.severity === monaco.MarkerSeverity.Error);

          if (errors.length === 0 && currentValue.trim()) {
            onRunRef.current();
          }
        }
      }
    );

    // Cleanup on unmount
    return () => {
      disposable.dispose();
      if (validateTimerRef.current) {
        clearTimeout(validateTimerRef.current);
      }
      editorRef.current?.dispose();
      editorRef.current = null;
      model.dispose();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
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

  const handleRunClick = useCallback(() => {
    // Check validation right before running - only block on errors, not warnings
    const model = editorRef.current?.getModel();
    if (!model) return;

    const markers = monaco.editor.getModelMarkers({ resource: model.uri });
    const errors = markers.filter(m => m.severity === monaco.MarkerSeverity.Error);

    if (errors.length === 0) {
      onRun();
    } else {
      setHasErrors(true);
      setErrorMessage(errors[0].message);
    }
  }, [onRun]);

  return (
    <div className="h-full flex flex-col bg-yapi-bg relative">
      {/* Editor Toolbar */}
      <div className="relative flex items-center justify-between px-6 h-16 border-b border-yapi-border/50 bg-yapi-bg-elevated/50 backdrop-blur-sm">
        {/* Subtle gradient accent */}
        <div className="absolute inset-0 bg-gradient-to-r from-yapi-accent/5 via-transparent to-transparent opacity-50"></div>

        <div className="relative flex items-center gap-4">
          <div className="flex items-center gap-2">
            <div className="w-1.5 h-1.5 rounded-full bg-yapi-accent shadow-[0_0_8px_rgba(255,102,0,0.5)] animate-pulse"></div>
            <h2 className="text-xs font-semibold text-yapi-fg tracking-wider">
              REQUEST
            </h2>
          </div>

          {hasErrors && (
            <div className="flex items-center gap-2 text-xs text-yapi-error bg-yapi-error/10 border border-yapi-error/20 px-3 py-1.5 rounded-lg backdrop-blur-sm" title={errorMessage}>
              <span className="font-medium">{errorMessage}</span>
            </div>
          )}
        </div>

        <button
          onClick={handleRunClick}
          disabled={hasErrors}
          className={`group relative px-5 py-2 text-sm font-semibold rounded-lg transition-all duration-300 flex items-center gap-2.5 overflow-hidden ${
            hasErrors
              ? "bg-yapi-bg-subtle text-yapi-fg-subtle cursor-not-allowed opacity-50"
              : "bg-gradient-to-r from-yapi-accent to-yapi-accent hover:from-yapi-accent hover:to-orange-500 text-white shadow-lg hover:shadow-xl hover:shadow-yapi-accent/30 hover:scale-105 active:scale-95"
          }`}
        >
          {!hasErrors && (
            <div className="absolute inset-0 bg-gradient-to-r from-white/0 via-white/20 to-white/0 opacity-0 group-hover:opacity-100 transition-opacity duration-500 rounded-lg animate-shimmer"></div>
          )}
          <span className="relative flex items-center gap-2">
            <span>Run</span>
            <kbd className="text-[10px] bg-black/30 px-1.5 py-0.5 rounded border border-white/10 font-mono">
              ⌘↵
            </kbd>
          </span>
        </button>
      </div>

      {/* Monaco Editor Container */}
      <div className="relative flex-1">
        <div ref={containerRef} className="h-full" />
      </div>

      <style>{`
        @keyframes shake {
          0%, 100% { transform: translateX(0); }
          25% { transform: translateX(-2px); }
          75% { transform: translateX(2px); }
        }

        @keyframes shimmer {
          0% { transform: translateX(-100%); }
          100% { transform: translateX(100%); }
        }

        .animate-shake {
          animation: shake 0.3s ease-in-out;
        }

        .animate-shimmer {
          animation: shimmer 2s ease-in-out infinite;
        }
      `}</style>
    </div>
  );
}
