"use client";

import * as monacoBase from "monaco-editor/esm/vs/editor/editor.api";

declare global {
  interface Window {
    MonacoEnvironment?: {
      getWorker: (moduleId: string, label: string) => Worker;
    };
  }
}

// Configure workers for monaco ESM + webpack/Next
if (typeof window !== "undefined") {
  window.MonacoEnvironment = {
    getWorker(_moduleId, label) {
      if (label === "json") {
        return new Worker(
          new URL(
            "monaco-editor/esm/vs/language/json/json.worker?worker",
            import.meta.url
          ),
          { type: "module" }
        );
      }

      if (label === "yaml") {
        return new Worker(
          new URL("monaco-yaml/yaml.worker", import.meta.url),
          { type: "module" }
        );
      }

      return new Worker(
        new URL(
          "monaco-editor/esm/vs/editor/editor.worker?worker",
          import.meta.url
        ),
        { type: "module" }
      );
    },
  };
}

export const monaco = monacoBase;
