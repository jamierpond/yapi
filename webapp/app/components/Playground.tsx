"use client";

import { useState, useEffect } from "react";
import { usePathname } from "next/navigation";
import OutputPanel from "./OutputPanel";
import type { ExecuteResponse } from "../types/api-contract";
import { yapiEncode, yapiDecode } from "../_lib/yapi-encode";

import dynamic from "next/dynamic";
const Editor = dynamic(() => import("./Editor"), { ssr: false });

const DEFAULT_YAML = `url: https://jsonplaceholder.typicode.com/posts
method: POST
content_type: application/json
query:
  userId: 1
  tags: example,demo

body:
  title: Example Post
  body: This is a more complex example with query parameters and a JSON body
  userId: 1
`;

export default function Playground() {
  const pathname = usePathname();
  const [yaml, setYaml] = useState(DEFAULT_YAML);
  const [result, setResult] = useState<ExecuteResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isInitialized, setIsInitialized] = useState(false);
  const [copyStatus, setCopyStatus] = useState<"idle" | "copied">("idle");

  // Load YAML from URL on mount
  useEffect(() => {
    if (typeof window === "undefined") return;

    const pathParts = pathname.split("/");
    if (pathParts[1] === "c" && pathParts[2]) {
      try {
        const decoded = yapiDecode(pathParts[2]);
        if (decoded) {
          setYaml(decoded);
        }
      } catch (e) {
        console.log("Failed to decode URL:", e);
      }
    }
    setIsInitialized(true);
  }, [pathname]);

  // Update URL when YAML changes using History API (no re-renders)
  useEffect(() => {
    if (!isInitialized || typeof window === "undefined") return;

    const encoded = yapiEncode(yaml);
    const newPath = `/c/${encoded}`;

    if (window.location.pathname !== newPath) {
      window.history.replaceState(null, "", newPath);
    }
  }, [yaml, isInitialized]);

  const handleYamlChange = (newYaml: string) => {
    setYaml(newYaml);
  };

  async function handleRun() {
    setIsLoading(true);
    setResult(null);

    try {
      const response = await fetch("/api/execute", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ yaml }),
      });

      const data = await response.json();
      setResult(data);
    } catch (error) {
      setResult({
        success: false,
        error: error instanceof Error ? error.message : "Unknown error occurred",
        errorType: "NETWORK_ERROR",
      });
    } finally {
      setIsLoading(false);
    }
  }

  async function handleShare() {
    try {
      await navigator.clipboard.writeText(window.location.href);
      setCopyStatus("copied");
      setTimeout(() => setCopyStatus("idle"), 2000);
    } catch (error) {
      console.error("Failed to copy URL:", error);
    }
  }

  return (
    <div className="flex flex-col h-screen bg-yapi-bg">
      {/* Header */}
      <header className="border-b border-yapi-border-dark bg-gradient-to-r from-yapi-header-from to-yapi-header-to">
        <div className="px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold font-mono text-yapi-fg">
                yapi playground
              </h1>
              <p className="text-sm text-yapi-fg/60 mt-1">
                compiler explorer for APIs
              </p>
            </div>
            <div className="flex items-center gap-3">
              <a
                href="https://github.com/jamierpond/yapi"
                target="_blank"
                rel="noopener noreferrer"
                className="text-yapi-fg/60 hover:text-yapi-fg transition-colors"
                aria-label="View source on GitHub"
              >
                <svg
                  width="24"
                  height="24"
                  viewBox="0 0 24 24"
                  fill="currentColor"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
                </svg>
              </a>
              <button
                onClick={handleShare}
                className="px-4 py-2 text-sm font-medium text-yapi-fg border border-yapi-border-dark rounded hover:bg-yellow-100 transition-colors"
              >
                {copyStatus === "copied" ? "copied!" : "share"}
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content - Split Pane */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left Panel - Editor */}
        <div className="w-1/2 border-r border-yapi-border-dark">
          <Editor value={yaml} onChange={handleYamlChange} onRun={handleRun} />
        </div>

        {/* Right Panel - Output */}
        <div className="w-1/2">
          <OutputPanel result={result} isLoading={isLoading} />
        </div>
      </div>
    </div>
  );
}
