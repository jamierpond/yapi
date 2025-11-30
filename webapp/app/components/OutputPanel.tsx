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
      <div className="h-full flex items-center justify-center bg-yapi-bg-subtle">
        <div className="flex flex-col items-center gap-4">
          <div className="w-10 h-10 border-2 border-yapi-accent border-t-transparent rounded-full animate-spin" />
          <p className="text-sm text-yapi-fg-muted">Executing request...</p>
        </div>
      </div>
    );
  }

  if (!result) {
    return (
      <div className="h-full flex items-center justify-center bg-yapi-bg-subtle">
        <div className="text-center text-yapi-fg-subtle">
          <p className="text-sm">Press ⌘↵ to run your API request</p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col bg-yapi-bg-subtle">
      {/* Response Section */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="flex items-center justify-between px-6 py-3 border-b border-yapi-border bg-yapi-bg-elevated">
          <h3 className="text-xs font-semibold text-yapi-fg-subtle uppercase tracking-wider">
            Response
          </h3>
          {isSuccessResponse(result) && (
            <div className="flex items-center gap-3">
              <span
                className={`text-xs font-mono px-2.5 py-1 rounded-md ${
                  result.statusCode >= 200 && result.statusCode < 300
                    ? "bg-yapi-success/10 text-yapi-success border border-yapi-success/20"
                    : result.statusCode >= 400
                    ? "bg-yapi-error/10 text-yapi-error border border-yapi-error/20"
                    : "bg-yapi-warning/10 text-yapi-warning border border-yapi-warning/20"
                }`}
              >
                {result.statusCode}
              </span>
              <span className="text-xs text-yapi-fg-muted">
                {result.timing}ms
              </span>
            </div>
          )}
        </div>

        {/* Content */}
        <div className="flex-1 overflow-auto p-6">
          {isSuccessResponse(result) ? (
            <pre className="text-sm text-yapi-fg whitespace-pre-wrap break-words">
              {JSON.stringify(result.responseBody, null, 2)}
            </pre>
          ) : (
            <div className="bg-yapi-error/5 border border-yapi-error/20 rounded-lg p-6">
              <div className="flex items-start gap-4">
                <span className="text-yapi-error text-xl">⚠</span>
                <div className="flex-1">
                  <h4 className="text-sm font-semibold text-yapi-error mb-2">
                    {result.errorType}
                  </h4>
                  <p className="text-sm text-yapi-fg-muted">{result.error}</p>
                  {!!result.details && (
                    <pre className="mt-3 text-xs text-yapi-fg-subtle bg-yapi-bg p-3 rounded-md overflow-x-auto border border-yapi-border">
                      {JSON.stringify(result.details, null, 2)}
                    </pre>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
