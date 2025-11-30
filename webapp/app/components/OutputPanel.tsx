"use client";

import type { ExecuteResponse } from "../types/api-contract";
import { isSuccessResponse } from "../types/api-contract";

interface OutputPanelProps {
  result: ExecuteResponse | null;
  isLoading: boolean;
}

export default function OutputPanel({ result, isLoading }: OutputPanelProps) {

  if (isLoading) {
    return (
      <div className="h-full flex items-center justify-center bg-yapi-bg-subtle relative overflow-hidden">
        {/* Animated gradient background */}
        <div className="absolute inset-0 bg-gradient-to-br from-yapi-accent/5 via-transparent to-orange-500/5 animate-pulse-slow"></div>

        <div className="relative flex flex-col items-center gap-6">
          {/* Multi-ring loader */}
          <div className="relative w-16 h-16">
            <div className="absolute inset-0 border-2 border-yapi-accent/30 border-t-yapi-accent rounded-full animate-spin"></div>
            <div className="absolute inset-2 border-2 border-orange-500/30 border-t-orange-500 rounded-full animate-spin-slow"></div>
            <div className="absolute inset-0 bg-yapi-accent/10 rounded-full animate-pulse"></div>
          </div>

          <div className="flex flex-col items-center gap-2">
            <p className="text-sm font-medium text-yapi-fg animate-pulse">Executing request</p>
            <div className="flex gap-1">
              <div className="w-1.5 h-1.5 bg-yapi-accent rounded-full animate-bounce"></div>
              <div className="w-1.5 h-1.5 bg-yapi-accent rounded-full animate-bounce animation-delay-150"></div>
              <div className="w-1.5 h-1.5 bg-yapi-accent rounded-full animate-bounce animation-delay-300"></div>
            </div>
          </div>
        </div>

        <style>{`
          @keyframes spin-slow {
            to { transform: rotate(-360deg); }
          }
          .animate-spin-slow {
            animation: spin-slow 2s linear infinite;
          }
          .animate-pulse-slow {
            animation: pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite;
          }
          .animation-delay-150 {
            animation-delay: 150ms;
          }
          .animation-delay-300 {
            animation-delay: 300ms;
          }
        `}</style>
      </div>
    );
  }

  if (!result) {
    return (
      <div className="h-full flex items-center justify-center bg-yapi-bg-subtle relative overflow-hidden">
        {/* Subtle animated background */}
        <div className="absolute inset-0 bg-gradient-to-br from-yapi-accent/3 via-transparent to-transparent"></div>

        <div className="relative text-center space-y-4">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-yapi-bg-elevated border border-yapi-border shadow-lg">
            <div className="text-2xl opacity-50">⚡</div>
          </div>
          <div className="space-y-2">
            <p className="text-sm text-yapi-fg-muted font-medium">Ready to execute</p>
            <p className="text-xs text-yapi-fg-subtle">
              Press <kbd className="px-2 py-1 text-[10px] bg-yapi-bg-elevated border border-yapi-border rounded font-mono">⌘↵</kbd> to run
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col bg-yapi-bg-subtle relative">
      {/* Response Section */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="relative flex items-center justify-between px-6 py-4 border-b border-yapi-border/50 bg-yapi-bg-elevated/50 backdrop-blur-sm">
          {/* Subtle gradient */}
          <div className="absolute inset-0 bg-gradient-to-r from-transparent via-yapi-accent/5 to-transparent opacity-50"></div>

          <div className="relative flex items-center gap-2">
            <div className="w-1.5 h-1.5 rounded-full bg-orange-500 shadow-[0_0_8px_rgba(249,115,22,0.5)] animate-pulse"></div>
            <h3 className="text-xs font-semibold text-yapi-fg tracking-wider">
              RESPONSE
            </h3>
          </div>

          {isSuccessResponse(result) && (
            <div className="relative flex items-center gap-4">
              <span
                className={`text-xs font-mono font-semibold px-3 py-1.5 rounded-lg backdrop-blur-sm transition-all duration-300 ${
                  result.statusCode >= 200 && result.statusCode < 300
                    ? "bg-gradient-to-r from-yapi-success/15 to-yapi-success/10 text-yapi-success border border-yapi-success/30 shadow-lg shadow-yapi-success/10"
                    : result.statusCode >= 400
                    ? "bg-gradient-to-r from-yapi-error/15 to-yapi-error/10 text-yapi-error border border-yapi-error/30 shadow-lg shadow-yapi-error/10"
                    : "bg-gradient-to-r from-yapi-warning/15 to-yapi-warning/10 text-yapi-warning border border-yapi-warning/30 shadow-lg shadow-yapi-warning/10"
                }`}
              >
                {result.statusCode}
              </span>
              <div className="flex items-center gap-2 px-3 py-1.5 bg-yapi-bg-elevated/50 border border-yapi-border/50 rounded-lg backdrop-blur-sm">
                <div className="w-1 h-1 rounded-full bg-yapi-accent animate-pulse"></div>
                <span className="text-xs text-yapi-fg-muted font-mono font-medium">
                  {result.timing}ms
                </span>
              </div>
            </div>
          )}
        </div>

        {/* Content */}
        <div className="flex-1 overflow-auto p-6 relative">
          {/* Subtle top gradient fade */}
          <div className="absolute top-0 left-0 right-0 h-16 bg-gradient-to-b from-yapi-bg-subtle via-yapi-bg-subtle/50 to-transparent pointer-events-none z-10"></div>

          <div className="relative">
            {isSuccessResponse(result) ? (
              <div className="group">
                <pre className="text-sm text-yapi-fg whitespace-pre-wrap break-words font-mono leading-relaxed p-6 bg-yapi-bg/30 border border-yapi-border/30 rounded-xl backdrop-blur-sm hover:border-yapi-border/50 transition-colors duration-300">
                  {JSON.stringify(result.responseBody, null, 2)}
                </pre>
              </div>
            ) : (
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
            )}
          </div>
        </div>
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
