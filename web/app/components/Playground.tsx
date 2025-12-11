"use client";

import { useState, useEffect } from "react";
import { usePathname } from "next/navigation";
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
  chain: {
    label: "Chain",
    yaml: `yapi: v1

# Mixed transport chain: HTTP -> gRPC -> HTTP
# Demonstrates variable references across transport types

chain:
  - name: get_todo
    url: https://jsonplaceholder.typicode.com/todos/1
    method: GET

  - name: grpc_hello
    url: grpc://grpcb.in:9000
    service: hello.HelloService
    rpc: SayHello
    plaintext: true
    body:
      greeting: $get_todo.title

  - name: create_post
    url: https://jsonplaceholder.typicode.com/posts
    method: POST
    headers:
      Content-Type: application/json
    body:
      original_todo: $get_todo.title
      grpc_reply: $grpc_hello.reply
      userId: $get_todo.userId
`,
  },
} as const;

type ExampleKey = keyof typeof EXAMPLES;

export default function Playground() {
  const pathname = usePathname();
  const [yaml, setYaml] = useState<string>(EXAMPLES.chain.yaml);
  const [result, setResult] = useState<ExecuteResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isInitialized, setIsInitialized] = useState(false);

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

  return (
    <div className="flex flex-col flex-1 bg-yapi-bg relative overflow-hidden">
      {/* Animated background orbs */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-0 -left-40 w-96 h-96 bg-yapi-accent/20 rounded-full mix-blend-screen filter blur-3xl opacity-30 animate-blob"></div>
        <div className="absolute top-0 -right-40 w-96 h-96 bg-orange-500/20 rounded-full mix-blend-screen filter blur-3xl opacity-30 animate-blob animation-delay-2000"></div>
        <div className="absolute -bottom-40 left-1/2 w-96 h-96 bg-purple-500/20 rounded-full mix-blend-screen filter blur-3xl opacity-20 animate-blob animation-delay-4000"></div>
      </div>

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

        .animate-blob {
          animation: blob 7s infinite;
        }

        .animation-delay-2000 {
          animation-delay: 2s;
        }

        .animation-delay-4000 {
          animation-delay: 4s;
        }
      `}</style>
    </div>
  );
}
