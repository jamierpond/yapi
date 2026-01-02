"use client";

import { useState } from "react";
import dynamic from "next/dynamic";
import type { ExecuteResponse } from "../types/api-contract";
import { isSuccessResponse } from "../types/api-contract";

const JsonViewer = dynamic(() => import("./JsonViewer"), { ssr: false });

interface OutputPanelProps {
  result: ExecuteResponse | null;
  isLoading: boolean;
}

type Tab = "body" | "headers" | "info" | "warnings";

export default function OutputPanel({ result, isLoading }: OutputPanelProps) {
  const [activeTab, setActiveTab] = useState<Tab>("body");

  if (isLoading) {
    return (
      <div className="h-full flex flex-col bg-yapi-bg relative">
        {/* Header skeleton that matches real header */}
        <div className="relative flex items-center justify-between px-6 h-16 border-b border-yapi-border/50 bg-yapi-bg-elevated/50 backdrop-blur-sm">
          <div className="absolute inset-0 bg-gradient-to-r from-yapi-accent/5 via-transparent to-transparent opacity-50"></div>

          <div className="relative flex items-center gap-2">
            <div className="w-1.5 h-1.5 rounded-full bg-yapi-border animate-pulse"></div>
            <div className="h-3 w-16 rounded-full yapi-skeleton"></div>
          </div>

          <div className="relative flex items-center gap-4">
            <div className="h-7 w-14 rounded-lg yapi-skeleton"></div>
            <div className="h-7 w-20 rounded-lg yapi-skeleton"></div>
          </div>
        </div>

        {/* Body skeleton that matches Monaco editor layout */}
        <div className="flex-1 overflow-hidden bg-yapi-bg relative">
          {/* Fake line numbers gutter - matches Monaco vs-dark background #1e1e1e */}
          <div className="absolute left-0 top-0 bottom-0 w-14 border-r border-yapi-border/30 flex flex-col items-center gap-[18px] pt-4 pb-4" style={{ backgroundColor: '#1e1e1e' }}>
            <div className="w-4 h-3 rounded yapi-skeleton opacity-40" />
            <div className="w-4 h-3 rounded yapi-skeleton opacity-40" />
            <div className="w-4 h-3 rounded yapi-skeleton opacity-40" />
            <div className="w-4 h-3 rounded yapi-skeleton opacity-40" />
            <div className="w-4 h-3 rounded yapi-skeleton opacity-40" />
            <div className="w-4 h-3 rounded yapi-skeleton opacity-40" />
            <div className="w-4 h-3 rounded yapi-skeleton opacity-40" />
            <div className="w-4 h-3 rounded yapi-skeleton opacity-40" />
          </div>

          {/* Fake code content - matches Monaco vs-dark background #1e1e1e */}
          <div className="absolute left-14 top-0 right-0 bottom-0 p-4 space-y-[18px]" style={{ backgroundColor: '#1e1e1e' }}>
            <div className="h-3 w-10/12 rounded yapi-skeleton" />
            <div className="h-3 w-9/12 rounded yapi-skeleton" />
            <div className="h-3 w-8/12 rounded yapi-skeleton" />
            <div className="h-3 w-7/12 rounded yapi-skeleton" />
            <div className="h-3 w-6/12 rounded yapi-skeleton" />
            <div className="h-3 w-9/12 rounded yapi-skeleton" />
            <div className="h-3 w-5/12 rounded yapi-skeleton" />
            <div className="h-3 w-4/12 rounded yapi-skeleton" />
          </div>
        </div>

        <style>{`
          @keyframes yapi-skeleton-shimmer {
            0% { background-position: -200% 0; }
            100% { background-position: 200% 0; }
          }

          .yapi-skeleton {
            background-image: linear-gradient(
              90deg,
              rgba(255,255,255,0.02) 0%,
              rgba(255,255,255,0.08) 20%,
              rgba(255,255,255,0.02) 40%
            );
            background-color: #1e1e1e;
            background-size: 200% 100%;
            animation: yapi-skeleton-shimmer 1.2s ease-in-out infinite;
          }
        `}</style>
      </div>
    );
  }

  if (!result) {
    return (
      <div className="h-full flex items-center justify-center bg-yapi-bg relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-yapi-accent/10 via-transparent to-transparent opacity-60"></div>

        <div className="relative text-center space-y-4">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-yapi-bg-elevated border border-yapi-border shadow-lg shadow-yapi-accent/10">
            <div className="text-2xl opacity-70">⚡</div>
          </div>
          <div className="space-y-2">
            <p className="text-sm text-yapi-fg-muted font-medium">
              Ready to execute
            </p>
            <p className="text-xs text-yapi-fg-subtle">
              Press{" "}
              <kbd className="px-2 py-1 text-[10px] bg-yapi-bg-elevated border border-yapi-border rounded font-mono">
                ⌘↵
              </kbd>{" "}
              to run
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col bg-yapi-bg relative">
      {/* Response Section */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="relative flex items-center justify-between px-6 h-16 border-b border-yapi-border/50 bg-yapi-bg-elevated/50 backdrop-blur-sm">
          {/* Shared gradient accent with editor */}
          <div className="absolute inset-0 bg-gradient-to-r from-yapi-accent/5 via-transparent to-transparent opacity-50"></div>

          <div className="relative flex items-center gap-2">
            <div className="w-1.5 h-1.5 rounded-full bg-yapi-accent shadow-[0_0_8px_rgba(255,102,0,0.5)] animate-pulse"></div>
            <h3 className="text-xs font-semibold text-yapi-fg tracking-wider">
              RESPONSE
            </h3>
          </div>

          {isSuccessResponse(result) && (
            <div className="relative flex items-center gap-4">
              <span
                className={`text-xs font-mono font-semibold px-3 py-1.5 rounded-lg backdrop-blur-sm ${
                  result.statusCode >= 200 && result.statusCode < 300
                    ? "bg-yapi-success/10 text-yapi-success border border-yapi-success/30"
                    : result.statusCode >= 400
                    ? "bg-yapi-error/10 text-yapi-error border border-yapi-error/30"
                    : "bg-yapi-warning/10 text-yapi-warning border border-yapi-warning/30"
                }`}
              >
                {result.statusCode}
              </span>
              <div className="flex items-center gap-2 px-3 py-1.5 bg-yapi-bg-elevated/70 border border-yapi-border/60 rounded-lg backdrop-blur-sm">
                <div className="w-1 h-1 rounded-full bg-yapi-accent animate-pulse"></div>
                <span className="text-xs text-yapi-fg-muted font-mono font-medium">
                  {result.timing}ms
                </span>
              </div>
            </div>
          )}
        </div>

        {/* Content */}
        {isSuccessResponse(result) ? (
          <div className="flex-1 flex flex-col overflow-hidden bg-yapi-bg">
            {/* Tabs */}
            <div className="flex items-center gap-1 px-4 pt-2 border-b border-yapi-border/30 bg-yapi-bg-elevated/30">
              <button
                onClick={() => setActiveTab("body")}
                className={`px-4 py-2 text-xs font-medium rounded-t-lg transition-all ${
                  activeTab === "body"
                    ? "bg-yapi-bg text-yapi-accent border-b-2 border-yapi-accent"
                    : "text-yapi-fg-muted hover:text-yapi-fg hover:bg-yapi-bg-elevated/50"
                }`}
              >
                Body
              </button>
              {result.headers && Object.keys(result.headers).length > 0 && (
                <button
                  onClick={() => setActiveTab("headers")}
                  className={`px-4 py-2 text-xs font-medium rounded-t-lg transition-all ${
                    activeTab === "headers"
                      ? "bg-yapi-bg text-yapi-accent border-b-2 border-yapi-accent"
                      : "text-yapi-fg-muted hover:text-yapi-fg hover:bg-yapi-bg-elevated/50"
                  }`}
                >
                  Headers
                  <span className="ml-2 px-1.5 py-0.5 text-[10px] bg-yapi-bg-elevated/70 rounded">
                    {Object.keys(result.headers).length}
                  </span>
                </button>
              )}
              <button
                onClick={() => setActiveTab("info")}
                className={`px-4 py-2 text-xs font-medium rounded-t-lg transition-all ${
                  activeTab === "info"
                    ? "bg-yapi-bg text-yapi-accent border-b-2 border-yapi-accent"
                    : "text-yapi-fg-muted hover:text-yapi-fg hover:bg-yapi-bg-elevated/50"
                }`}
              >
                Info
              </button>
              {result.warnings && result.warnings.length > 0 && (
                <button
                  onClick={() => setActiveTab("warnings")}
                  className={`px-4 py-2 text-xs font-medium rounded-t-lg transition-all ${
                    activeTab === "warnings"
                      ? "bg-yapi-bg text-yapi-accent border-b-2 border-yapi-accent"
                      : "text-yapi-fg-muted hover:text-yapi-fg hover:bg-yapi-bg-elevated/50"
                  }`}
                >
                  Warnings
                  <span className="ml-2 px-1.5 py-0.5 text-[10px] bg-yapi-warning/20 text-yapi-warning rounded">
                    {result.warnings.length}
                  </span>
                </button>
              )}
            </div>

            {/* Tab Content */}
            <div className="flex-1 overflow-hidden">
              {activeTab === "body" && (
                <JsonViewer
                  value={
                    typeof result.responseBody === 'string'
                      ? result.responseBody
                      : JSON.stringify(result.responseBody, null, 2)
                  }
                />
              )}

              {activeTab === "headers" && result.headers && (
                <div className="h-full overflow-auto p-6">
                  <div className="space-y-2">
                    {Object.entries(result.headers).map(([key, value]) => (
                      <div
                        key={key}
                        className="flex items-start gap-4 p-3 bg-yapi-bg-elevated/50 border border-yapi-border/30 rounded-lg hover:border-yapi-border/60 transition-colors"
                      >
                        <span className="text-xs font-mono font-semibold text-yapi-accent min-w-[140px]">
                          {key}:
                        </span>
                        <span className="text-xs font-mono text-yapi-fg-muted flex-1 break-all">
                          {value}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {activeTab === "info" && (
                <div className="h-full overflow-auto p-6">
                  <div className="space-y-4">
                    {/* Transport type */}
                    {result.transport && (
                      <div className="flex items-center gap-4 p-3 bg-yapi-bg-elevated/50 border border-yapi-border/30 rounded-lg">
                        <span className="text-xs font-semibold text-yapi-fg-muted min-w-[120px]">
                          Transport
                        </span>
                        <span className="px-2 py-1 text-xs font-mono font-semibold bg-yapi-accent/10 text-yapi-accent border border-yapi-accent/30 rounded">
                          {result.transport.toUpperCase()}
                        </span>
                      </div>
                    )}

                    {/* Request URL */}
                    {result.requestUrl && (
                      <div className="flex items-start gap-4 p-3 bg-yapi-bg-elevated/50 border border-yapi-border/30 rounded-lg">
                        <span className="text-xs font-semibold text-yapi-fg-muted min-w-[120px]">
                          URL
                        </span>
                        <span className="text-xs font-mono text-yapi-fg break-all flex-1">
                          {result.requestUrl}
                        </span>
                      </div>
                    )}

                    {/* Method */}
                    {result.method && (
                      <div className="flex items-center gap-4 p-3 bg-yapi-bg-elevated/50 border border-yapi-border/30 rounded-lg">
                        <span className="text-xs font-semibold text-yapi-fg-muted min-w-[120px]">
                          Method
                        </span>
                        <span className="px-2 py-1 text-xs font-mono font-semibold bg-blue-500/10 text-blue-400 border border-blue-500/30 rounded">
                          {result.method}
                        </span>
                      </div>
                    )}

                    {/* Service (gRPC) */}
                    {result.service && (
                      <div className="flex items-start gap-4 p-3 bg-yapi-bg-elevated/50 border border-yapi-border/30 rounded-lg">
                        <span className="text-xs font-semibold text-yapi-fg-muted min-w-[120px]">
                          Service
                        </span>
                        <span className="text-xs font-mono text-yapi-fg flex-1">
                          {result.service}
                        </span>
                      </div>
                    )}

                    {/* Status Code */}
                    {result.statusCode !== undefined && (
                      <div className="flex items-center gap-4 p-3 bg-yapi-bg-elevated/50 border border-yapi-border/30 rounded-lg">
                        <span className="text-xs font-semibold text-yapi-fg-muted min-w-[120px]">
                          Status
                        </span>
                        <span
                          className={`px-2 py-1 text-xs font-mono font-semibold rounded ${
                            result.statusCode >= 200 && result.statusCode < 300
                              ? "bg-yapi-success/10 text-yapi-success border border-yapi-success/30"
                              : result.statusCode >= 400
                              ? "bg-yapi-error/10 text-yapi-error border border-yapi-error/30"
                              : "bg-yapi-warning/10 text-yapi-warning border border-yapi-warning/30"
                          }`}
                        >
                          {result.statusCode}
                        </span>
                      </div>
                    )}

                    {/* Timing */}
                    <div className="flex items-center gap-4 p-3 bg-yapi-bg-elevated/50 border border-yapi-border/30 rounded-lg">
                      <span className="text-xs font-semibold text-yapi-fg-muted min-w-[120px]">
                        Time
                      </span>
                      <span className="text-xs font-mono text-yapi-fg">
                        {result.timing}ms
                      </span>
                    </div>

                    {/* Size */}
                    {result.sizeBytes !== undefined && (
                      <div className="flex items-center gap-4 p-3 bg-yapi-bg-elevated/50 border border-yapi-border/30 rounded-lg">
                        <span className="text-xs font-semibold text-yapi-fg-muted min-w-[120px]">
                          Size
                        </span>
                        <span className="text-xs font-mono text-yapi-fg">
                          {result.sizeBytes} bytes
                          {result.sizeLines !== undefined && ` • ${result.sizeLines} lines`}
                          {result.sizeChars !== undefined && ` • ${result.sizeChars} chars`}
                        </span>
                      </div>
                    )}

                    {/* Content Type */}
                    {result.contentType && (
                      <div className="flex items-start gap-4 p-3 bg-yapi-bg-elevated/50 border border-yapi-border/30 rounded-lg">
                        <span className="text-xs font-semibold text-yapi-fg-muted min-w-[120px]">
                          Content-Type
                        </span>
                        <span className="text-xs font-mono text-yapi-fg flex-1 break-all">
                          {result.contentType}
                        </span>
                      </div>
                    )}
                  </div>
                </div>
              )}

              {activeTab === "warnings" && result.warnings && (
                <div className="h-full overflow-auto p-6">
                  <div className="space-y-3">
                    {result.warnings.map((warning, idx) => (
                      <div
                        key={idx}
                        className="flex items-start gap-3 p-4 bg-yapi-warning/5 border border-yapi-warning/20 rounded-lg"
                      >
                        <span className="text-yapi-warning text-sm flex-shrink-0 mt-0.5">⚠</span>
                        <p className="text-sm text-yapi-fg-muted leading-relaxed flex-1">
                          {warning}
                        </p>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        ) : (
          <div className="flex-1 overflow-hidden bg-yapi-bg">
            <div className="h-full w-full px-6 py-4">
              <div className="relative overflow-hidden rounded-xl bg-gradient-to-br from-yapi-error/10 via-yapi-error/5 to-transparent border border-yapi-error/30 p-6 backdrop-blur-sm animate-error-pulse">
                {/* Error glow effect */}
                <div className="absolute top-0 right-0 w-32 h-32 bg-yapi-error/20 rounded-full blur-3xl"></div>

                <div className="relative flex items-start gap-4">
                  <div className="flex-shrink-0 w-10 h-10 rounded-full bg-yapi-error/20 border border-yapi-error/30 flex items-center justify-center">
                    <span className="text-yapi-error text-lg">⚠</span>
                  </div>
                  <div className="flex-1 space-y-3">
                    <div className="flex items-center gap-2">
                      <div className="h-px flex-1 bg-gradient-to-r from-yapi-error/30 to-transparent"></div>
                    </div>
                    <h4 className="text-sm font-bold text-yapi-error tracking-wide">
                      {result.errorType}
                    </h4>
                    <p className="text-sm text-yapi-fg leading-relaxed">{result.error}</p>
                    {!!result.details && (
                      <div className="mt-4 p-4 bg-yapi-bg/50 border border-yapi-border/50 rounded-lg backdrop-blur-sm">
                        <pre className="text-xs text-yapi-fg-subtle font-mono overflow-x-auto leading-relaxed">
                          {JSON.stringify(result.details, null, 2)}
                        </pre>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      <style>{`
        @keyframes error-pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.95; }
        }
        .animate-error-pulse {
          animation: error-pulse 2s ease-in-out infinite;
        }
      `}</style>
    </div>
  );
}
