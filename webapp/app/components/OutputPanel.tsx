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
      <div className="h-full flex items-center justify-center bg-yapi-output">
        <div className="flex flex-col items-center gap-3">
          <div className="w-8 h-8 border-3 border-yapi-accent border-t-transparent rounded-full animate-spin" />
          <p className="text-sm text-yapi-fg/60 font-mono">Executing request...</p>
        </div>
      </div>
    );
  }

  if (!result) {
    return (
      <div className="h-full flex items-center justify-center bg-yapi-output">
        <div className="text-center text-yapi-fg/40">
          <p className="text-sm font-mono">Press ⌘↵ to run your API request</p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col bg-yapi-output">
      {/* Response Section */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="flex items-center gap-4 px-4 py-2 border-b border-yapi-border bg-yellow-50">
          <h3 className="text-xs font-semibold text-yapi-fg/60 uppercase tracking-wide">
            Response
          </h3>
          {isSuccessResponse(result) && (
            <div className="flex items-center gap-3 ml-auto">
              <span
                className={`text-xs font-mono px-2 py-0.5 rounded ${
                  result.statusCode >= 200 && result.statusCode < 300
                    ? "bg-green-100 text-green-700"
                    : result.statusCode >= 400
                    ? "bg-red-100 text-red-700"
                    : "bg-orange-100 text-orange-700"
                }`}
              >
                {result.statusCode}
              </span>
              <span className="text-xs text-yapi-fg/60 font-mono">
                {result.timing}ms
              </span>
            </div>
          )}
        </div>

        {/* Content */}
        <div className="flex-1 overflow-auto p-4">
          {isSuccessResponse(result) ? (
            <pre className="text-sm font-mono text-yapi-fg whitespace-pre-wrap break-words">
              {JSON.stringify(result.responseBody, null, 2)}
            </pre>
          ) : (
            <div className="bg-red-50 border border-red-200 rounded p-4">
              <div className="flex items-start gap-3">
                <span className="text-red-500 text-lg">⚠️</span>
                <div className="flex-1">
                  <h4 className="text-sm font-semibold text-red-700 mb-1">
                    {result.errorType}
                  </h4>
                  <p className="text-sm text-red-600">{result.error}</p>
                  {!!result.details && (
                    <pre className="mt-2 text-xs font-mono text-red-500 whitespace-pre-wrap">
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
