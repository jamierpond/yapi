"use client";

import { useState, useEffect } from "react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import OutputPanel from "./OutputPanel";
import type { ExecuteResponse } from "../types/api-contract";
import { yapiEncode, yapiDecode } from "../_lib/yapi-encode";

import dynamic from "next/dynamic";
const Editor = dynamic(() => import("./Editor"), { ssr: false });

const EXAMPLES = {
  http: {
    label: "HTTP",
    yaml: `yapi: v1
# POST request with JSON body
url: https://httpbin.org/post
method: POST
content_type: application/json

body:
  title: "Hello from yapi"
  content: "This is a test post"
  tags:
    - testing
    - api
  metadata:
    source: yapi
    version: "1.0"
`,
  },
  graphql: {
    label: "GraphQL",
    yaml: `yapi: v1
# GraphQL - List all countries
url: https://countries.trevorblades.com/graphql

graphql: |
  query {
    countries {
      code
      name
      capital
      currency
    }
  }

jq_filter: ".data.countries[:5]"
`,
  },
  grpc: {
    label: "gRPC",
    yaml: `yapi: v1
# gRPC Hello Service example
# Uses grpcb.in test service with server reflection
url: grpc://grpcb.in:9000

service: hello.HelloService
rpc: SayHello
plaintext: true

body:
  greeting: "World"
`,
  },
  tcp: {
    label: "TCP",
    yaml: `yapi: v1
# TCP echo server test
url: tcp://tcpbin.com:4242

data: "Hello from yapi!\\n"
encoding: text
read_timeout: 5
idle_timeout: 500
close_after_send: true
`,
  },
} as const;

type ExampleKey = keyof typeof EXAMPLES;

export default function Playground() {
  const pathname = usePathname();
  const [yaml, setYaml] = useState<string>(EXAMPLES.http.yaml);
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

  const handleLoadExample = (key: ExampleKey) => {
    setYaml(EXAMPLES[key].yaml);
    setResult(null);
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
    <div className="flex flex-col h-screen bg-yapi-bg relative overflow-hidden">
      {/* Animated background orbs */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-0 -left-40 w-96 h-96 bg-yapi-accent/20 rounded-full mix-blend-screen filter blur-3xl opacity-30 animate-blob"></div>
        <div className="absolute top-0 -right-40 w-96 h-96 bg-orange-500/20 rounded-full mix-blend-screen filter blur-3xl opacity-30 animate-blob animation-delay-2000"></div>
        <div className="absolute -bottom-40 left-1/2 w-96 h-96 bg-purple-500/20 rounded-full mix-blend-screen filter blur-3xl opacity-20 animate-blob animation-delay-4000"></div>
      </div>

      {/* Header */}
      <header className="relative border-b border-yapi-border-strong/50 backdrop-blur-2xl overflow-hidden z-10">
        {/* Noise texture overlay */}
        <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIzMDAiIGhlaWdodD0iMzAwIj48ZmlsdGVyIGlkPSJhIiB4PSIwIiB5PSIwIj48ZmVUdXJidWxlbmNlIGJhc2VGcmVxdWVuY3k9Ii43NSIgc3RpdGNoVGlsZXM9InN0aXRjaCIgdHlwZT0iZnJhY3RhbE5vaXNlIi8+PGZlQ29sb3JNYXRyaXggdHlwZT0ic2F0dXJhdGUiIHZhbHVlcz0iMCIvPjwvZmlsdGVyPjxwYXRoIGQ9Ik0wIDBoMzAwdjMwMEgweiIgZmlsdGVyPSJ1cmwoI2EpIiBvcGFjaXR5PSIuMDUiLz48L3N2Zz4=')] opacity-40"></div>

        {/* Kinetic gradient */}
        <div className="absolute inset-0 bg-gradient-to-r from-yapi-accent/10 via-transparent to-orange-500/10 opacity-60"></div>
        <div className="absolute inset-0 bg-gradient-to-b from-yapi-bg-elevated/95 via-yapi-bg-elevated/90 to-yapi-bg-elevated/95"></div>

        {/* Glow line at bottom */}
        <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-yapi-accent/50 to-transparent"></div>

        <div className="relative max-w-[1800px] mx-auto px-8 py-6">
          <div className="flex items-center justify-between">
            {/* Left: Logo & Title */}
            <div className="flex items-center gap-5">
              <Link href="/" className="relative group cursor-pointer">
                <div className="absolute inset-0 bg-yapi-accent/20 blur-xl group-hover:blur-2xl transition-all duration-500 rounded-full"></div>
                <div className="relative text-5xl transform group-hover:scale-110 transition-transform duration-300">🐑</div>
              </Link>
              <div className="space-y-1">
                <h1 className="text-2xl font-bold tracking-tight">
                  <span className="bg-gradient-to-r from-white via-yapi-accent to-orange-400 bg-clip-text text-transparent bg-[length:200%_auto] animate-gradient-shift">
                    yapi playground
                  </span>
                </h1>
                <p className="text-xs text-yapi-fg-subtle tracking-wide font-light">
                  compiler explorer for APIs
                </p>
              </div>
            </div>

            {/* Right: Actions */}
            <div className="flex items-center gap-4">
              <a
                href="https://github.com/jamierpond/yapi"
                target="_blank"
                rel="noopener noreferrer"
                className="group flex items-center gap-2.5 px-4 py-2.5 text-sm text-yapi-fg-muted hover:text-yapi-fg rounded-lg hover:bg-yapi-bg-subtle/50 transition-all duration-300 backdrop-blur-sm border border-transparent hover:border-yapi-border"
                aria-label="View source on GitHub"
              >
                <svg
                  width="18"
                  height="18"
                  viewBox="0 0 24 24"
                  fill="currentColor"
                  xmlns="http://www.w3.org/2000/svg"
                  className="group-hover:rotate-12 transition-transform duration-300"
                >
                  <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
                </svg>
                <span className="font-medium">source</span>
              </a>

              <button
                onClick={handleShare}
                className="group relative px-5 py-2.5 text-sm font-semibold text-white rounded-lg overflow-hidden transition-all duration-300 hover:scale-105 active:scale-95"
              >
                <div className="absolute inset-0 bg-gradient-to-r from-yapi-accent via-orange-500 to-yapi-accent bg-[length:200%_auto] animate-gradient-shift"></div>
                <div className="absolute inset-0 bg-gradient-to-r from-yapi-accent/0 via-white/20 to-yapi-accent/0 opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
                <span className="relative flex items-center gap-2">
                  {copyStatus === "copied" ? (
                    <>
                      <span className="inline-block animate-bounce-in">✓</span>
                      <span>copied</span>
                    </>
                  ) : (
                    "share"
                  )}
                </span>
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content - Split Pane */}
      <div className="flex-1 flex overflow-hidden relative z-0">
        {/* Left Panel - Editor */}
        <div className="w-1/2 relative group z-10">
          <div className="absolute -right-px top-0 bottom-0 w-px bg-gradient-to-b from-transparent via-yapi-border-strong to-transparent"></div>
          <Editor
            value={yaml}
            onChange={handleYamlChange}
            onRun={handleRun}
            examples={Object.entries(EXAMPLES).map(([key, { label, yaml }]) => ({
              key,
              label,
              defaultYaml: yaml,
            }))}
            onLoadExample={(key) => handleLoadExample(key as ExampleKey)}
          />
        </div>

        {/* Right Panel - Output */}
        <div className="w-1/2 relative">
          <OutputPanel result={result} isLoading={isLoading} />
        </div>
      </div>

      <style>{`
        @keyframes blob {
          0%, 100% { transform: translate(0, 0) scale(1); }
          33% { transform: translate(30px, -50px) scale(1.1); }
          66% { transform: translate(-20px, 20px) scale(0.9); }
        }

        @keyframes gradient-shift {
          0% { background-position: 0% 50%; }
          50% { background-position: 100% 50%; }
          100% { background-position: 0% 50%; }
        }

        @keyframes bounce-in {
          0% { transform: scale(0); }
          50% { transform: scale(1.2); }
          100% { transform: scale(1); }
        }

        .animate-blob {
          animation: blob 7s infinite;
        }

        .animation-delay-2000 {
          animation-delay: 2s;
        }

        .animation-delay-4000 {
          animation-delay: 4s;
        }

        .animate-gradient-shift {
          animation: gradient-shift 3s ease infinite;
        }

        .animate-bounce-in {
          animation: bounce-in 0.3s ease-out;
        }
      `}</style>
    </div>
  );
}
