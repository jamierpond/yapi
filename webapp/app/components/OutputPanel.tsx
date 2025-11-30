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
      <div className="h-full flex flex-col bg-yapi-bg relative">
        {/* Header skeleton */}
        <div className="relative flex items-center justify-between px-6 h-16 border-b border-yapi-border/50 bg-yapi-bg-elevated/50 backdrop-blur-sm">
          <div className="flex items-center gap-2">
            <div className="w-1.5 h-1.5 rounded-full bg-yapi-border animate-pulse"></div>
            <div className="h-3 w-20 bg-yapi-border rounded animate-shimmer-skeleton"></div>
          </div>
          <div className="flex items-center gap-4">
            <div className="h-7 w-16 bg-yapi-border/50 rounded-lg animate-shimmer-skeleton"></div>
            <div className="h-7 w-20 bg-yapi-border/50 rounded-lg animate-shimmer-skeleton"></div>
          </div>
        </div>

        {/* Content skeleton */}
        <div className="flex-1 overflow-hidden p-6 bg-yapi-bg-subtle/30">
          <div className="space-y-3">
            <div className="h-4 bg-yapi-border/40 rounded animate-shimmer-skeleton"></div>
            <div className="h-4 bg-yapi-border/40 rounded w-5/6 animate-shimmer-skeleton animation-delay-100"></div>
            <div className="h-4 bg-yapi-border/40 rounded w-4/6 animate-shimmer-skeleton animation-delay-200"></div>
            <div className="h-4 bg-yapi-border/40 rounded w-3/4 animate-shimmer-skeleton animation-delay-100"></div>
            <div className="h-4 bg-yapi-border/40 rounded w-2/3 animate-shimmer-skeleton animation-delay-200"></div>
            <div className="h-4 bg-yapi-border/40 rounded animate-shimmer-skeleton"></div>
            <div className="h-4 bg-yapi-border/40 rounded w-4/5 animate-shimmer-skeleton animation-delay-100"></div>
            <div className="h-4 bg-yapi-border/40 rounded w-3/5 animate-shimmer-skeleton animation-delay-200"></div>
          </div>
        </div>

        <style>{`
          @keyframes shimmer-skeleton {
            0% { opacity: 0.4; }
            50% { opacity: 1; }
            100% { opacity: 0.4; }
          }
          .animate-shimmer-skeleton {
            animation: shimmer-skeleton 1.5s ease-in-out infinite;
            background: linear-gradient(
              90deg,
              rgba(255, 255, 255, 0.05) 0%,
              rgba(255, 255, 255, 0.15) 50%,
              rgba(255, 255, 255, 0.05) 100%
            );
            background-size: 200% 100%;
            animation: shimmer-skeleton 1.5s ease-in-out infinite;
          }
          @keyframes shimmer-skeleton {
            0% { background-position: -200% 0; }
            100% { background-position: 200% 0; }
          }
          .animation-delay-100 {
            animation-delay: 100ms;
          }
          .animation-delay-200 {
            animation-delay: 200ms;
          }
        `}</style>
      </div>
    );
  }

  if (!result) {
    return (
      <div className="h-full flex items-center justify-center bg-yapi-bg relative overflow-hidden">
        {/* Subtle animated background */}
        <div className="absolute inset-0 bg-gradient-to-br from-orange-500/3 via-transparent to-transparent"></div>

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
    <div className="h-full flex flex-col bg-yapi-bg relative">
      {/* Response Section */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="relative flex items-center justify-between px-6 h-16 border-b border-yapi-border/50 bg-yapi-bg-elevated/50 backdrop-blur-sm">
          {/* Subtle gradient accent */}
          <div className="absolute inset-0 bg-gradient-to-r from-transparent via-orange-500/5 to-transparent opacity-50"></div>

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
        <div className="flex-1 overflow-auto p-6 relative bg-yapi-bg-subtle/30">
          {/* Subtle top gradient fade */}
          <div className="absolute top-0 left-0 right-0 h-16 bg-gradient-to-b from-yapi-bg-subtle/30 via-yapi-bg-subtle/15 to-transparent pointer-events-none z-10"></div>

          <div className="relative">
            {isSuccessResponse(result) ? (
              <div className="group">
                <pre className="text-sm text-yapi-fg whitespace-pre-wrap break-words font-mono leading-relaxed p-6 bg-yapi-bg-elevated/40 border border-yapi-border/30 rounded-xl backdrop-blur-sm hover:border-yapi-border/50 transition-colors duration-300">
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
