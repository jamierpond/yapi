"use client";

import * as monacoBase from "monaco-editor/esm/vs/editor/editor.api";

// languages / contribs
import "monaco-editor/esm/vs/basic-languages/yaml/yaml.contribution";
import "monaco-editor/esm/vs/language/json/monaco.contribution";

// no declare-global here

if (typeof window !== "undefined") {
  (window as any).MonacoEnvironment = {
    getWorker(_moduleId: string, label: string) {
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

