// app/lib/monaco.ts
"use client";

import * as monacoBase from "monaco-editor/esm/vs/editor/editor.api";

// 1) Register YAML language so you get colors & tokenization
import "monaco-editor/esm/vs/basic-languages/yaml/yaml.contribution";

// 2) Make sure core editor contribs needed for LSP UX are present
//    (suggest, hover, etc). With the ESM build these are *not* all
//    guaranteed unless you import them somewhere.
import "monaco-editor/esm/vs/editor/contrib/suggest/browser/suggestController";

declare global {
  interface Window {
    MonacoEnvironment?: {
      getWorker: (moduleId: string, label: string) => Worker;
    };
  }
}

// 3) Workers – your structure is fine, just follow the monaco-yaml docs
//    and don't use ?worker here; new URL(..., import.meta.url) is enough.
if (typeof window !== "undefined") {
  window.MonacoEnvironment = {
    getWorker(_moduleId, label) {
      if (label === "json") {
        return new Worker(
          new URL(
            "monaco-editor/esm/vs/language/json/json.worker",
            import.meta.url
          ),
          { type: "module" }
        );
      }

      if (label === "yaml") {
        // monaco-yaml worker
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

export const monaco = monacoBase;

