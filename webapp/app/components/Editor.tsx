"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import * as monaco from "monaco-editor";
import { configureMonacoYaml } from "monaco-yaml";
import yaml from "yaml";

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

    // Configure workers once; Monaco uses this to spawn background workers
    if (typeof window !== "undefined" && !(window as any).MonacoEnvironment) {
      (window as any).MonacoEnvironment = {
        getWorker(_id: string, label: string) {
          // Use YAML worker for yaml language
          if (label === "yaml") {
            return new Worker(
              new URL("monaco-yaml/yaml.worker", import.meta.url),
              { type: "module" }
            );
          }

          // Fallback to default Monaco editor worker for everything else
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
    <div className="h-full flex flex-col bg-yapi-bg">
      {/* Editor Toolbar */}
      <div className="flex items-center justify-between px-6 py-3 border-b border-yapi-border bg-yapi-bg-elevated">
        <div className="flex items-center gap-3">
          <h2 className="text-xs font-semibold text-yapi-fg-subtle uppercase tracking-wider">
            Request Config
          </h2>
          {hasErrors && (
            <div className="flex items-center gap-2 text-xs text-yapi-error bg-yapi-error/10 px-2 py-1 rounded-md">
              <span>⚠</span>
              <span>{errorMessage}</span>
            </div>
          )}
        </div>
        <button
          onClick={handleRunClick}
          disabled={hasErrors}
          className={`px-5 py-2 text-sm font-medium rounded-md transition-all flex items-center gap-2 ${
            hasErrors
              ? "bg-yapi-bg-subtle text-yapi-fg-subtle cursor-not-allowed"
              : "bg-yapi-accent text-white hover:bg-yapi-accent-hover shadow-md hover:shadow-glow"
          }`}
        >
          <span>Run</span>
          <kbd className="text-xs bg-black/20 px-2 py-0.5 rounded">⌘↵</kbd>
        </button>
      </div>

      {/* Monaco Editor Container */}
      <div ref={containerRef} className="flex-1" />
    </div>
  );
}
