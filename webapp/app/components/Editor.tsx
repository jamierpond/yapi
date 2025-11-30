"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import type * as Monaco from "monaco-editor";      // types only
import { monaco } from "../lib/monaco";            // runtime API
import { configureMonacoYaml } from "monaco-yaml";

// Simple guard so we only wire YAML services once
let yamlConfigured = false;

interface EditorProps {
  value: string;
  onChange: (value: string) => void;
  onRun: () => void;
}

export default function Editor({ value, onChange, onRun }: EditorProps) {
  // Ref to the DOM node Monaco will render into
  const containerRef = useRef<HTMLDivElement | null>(null);
  // Ref to keep the editor instance (avoid double init, allow dispose)
  const editorRef = useRef<Monaco.editor.IStandaloneCodeEditor | null>(null);
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

  // Simple validation: check Monaco's markers
  const checkValidation = useCallback(() => {
    const model = editorRef.current?.getModel();
    if (!model) {
      setHasErrors(false);
      setErrorMessage("");
      return;
    }

    // Get Monaco's validation markers
    const markers = monaco.editor.getModelMarkers({ resource: model.uri });
    const problems = markers.filter(
      m => m.severity === monaco.MarkerSeverity.Error ||
           m.severity === monaco.MarkerSeverity.Warning
    );

    if (problems.length > 0) {
      setHasErrors(true);
      setErrorMessage(problems[0].message);
    } else {
      setHasErrors(false);
      setErrorMessage("");
    }
  }, []);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    // If React re-runs the effect (StrictMode), do not re-create the editor
    if (editorRef.current) return;

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

    // Configure YAML validation / completion once per app
    if (!yamlConfigured) {
      configureMonacoYaml(monaco, {
        // Basic LSP-ish features
        validate: true,
        hover: true,
        completion: true,
        format: true,
        enableSchemaRequest: true,

        // Here you wire your JSON schema(s)
        schemas: [
          {
            uri: "https://pond.audio/yapi/schema",
            fileMatch: ["*"],
          },
        ],
      });

      yamlConfigured = true;
    }

    // Listen to content changes
    const disposable = editorRef.current.onDidChangeModelContent(() => {
      console.log("Content changed");
      const currentValue = editorRef.current?.getValue() || "";
      onChange(currentValue);
    });

    // Listen to Monaco marker changes (validation updates)
    const markerDisposable = monaco.editor.onDidChangeMarkers((uris) => {
      console.log("Markers changed for URIs:", uris);
      const model = editorRef.current?.getModel();
      if (!model) return;

      // Check if markers changed for our model
      if (uris.some(uri => uri.toString() === model.uri.toString())) {
        checkValidation();
      }
    });

    // Add keyboard shortcut for Cmd+Enter (Mac) or Ctrl+Enter (Windows/Linux)
    editorRef.current.addCommand(
      monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter,
      () => {
        // Get current value and validate
        const currentValue = editorRef.current?.getValue() || "";
        const model = editorRef.current?.getModel();

        if (model) {
          const markers = monaco.editor.getModelMarkers({ resource: model.uri });
          const problems = markers.filter(
            m => m.severity === monaco.MarkerSeverity.Error ||
                 m.severity === monaco.MarkerSeverity.Warning
          );

          // Only run if no problems
          if (problems.length === 0 && currentValue.trim()) {
            onRunRef.current();
          }
        }
      }
    );

    // Add keyboard shortcut for Cmd+S (Mac) or Ctrl+S (Windows/Linux)
    editorRef.current.addCommand(
      monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS,
      () => {
        // Get current value and validate
        const currentValue = editorRef.current?.getValue() || "";
        const model = editorRef.current?.getModel();

        if (model) {
          const markers = monaco.editor.getModelMarkers({ resource: model.uri });
          const problems = markers.filter(
            m => m.severity === monaco.MarkerSeverity.Error ||
                 m.severity === monaco.MarkerSeverity.Warning
          );

          // Only run if no problems
          if (problems.length === 0 && currentValue.trim()) {
            onRunRef.current();
          }
        }
      }
    );

    // Cleanup on unmount
    return () => {
      disposable.dispose();
      markerDisposable.dispose();
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
    // Check validation right before running
    const model = editorRef.current?.getModel();
    if (!model) return;

    const markers = monaco.editor.getModelMarkers({ resource: model.uri });
    const problems = markers.filter(
      m => m.severity === monaco.MarkerSeverity.Error ||
           m.severity === monaco.MarkerSeverity.Warning
    );

    // Only run if there are no validation problems
    if (problems.length === 0) {
      onRun();
    } else {
      // Update error state to show the user why it didn't run
      setHasErrors(true);
      setErrorMessage(problems[0].message);
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
            <div className="group relative flex items-center gap-2 text-xs text-yapi-error bg-yapi-error/10 border border-yapi-error/20 px-3 py-1.5 rounded-lg backdrop-blur-sm animate-shake">
              <span className="text-sm">⚠</span>
              <span className="font-medium max-w-xs truncate">{errorMessage}</span>

              {/* Tooltip on hover */}
              <div className="absolute left-0 top-full mt-2 hidden group-hover:block z-50">
                <div className="bg-yapi-bg-elevated border border-yapi-error/30 rounded-lg px-3 py-2 shadow-xl max-w-md">
                  <p className="text-xs text-yapi-fg whitespace-pre-wrap">{errorMessage}</p>
                </div>
              </div>
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

      {/* Monaco Editor Container with subtle inner glow */}
      <div className="relative flex-1 overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-b from-yapi-accent/5 via-transparent to-transparent pointer-events-none h-32"></div>
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
