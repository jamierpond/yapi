"use client";

import type { AssertionResult } from "../../types/api-contract";

interface TestResultItemProps {
  assertion: AssertionResult;
  index: number;
}

function StatusIcon({ passed }: { passed: boolean }) {
  if (passed) {
    return (
      <span className="flex items-center justify-center w-6 h-6 rounded-full bg-yapi-success/20 text-yapi-success shrink-0">
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M5 13l4 4L19 7" />
        </svg>
      </span>
    );
  }
  return (
    <span className="flex items-center justify-center w-6 h-6 rounded-full bg-yapi-error/20 text-yapi-error shrink-0">
      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M6 18L18 6M6 6l12 12" />
      </svg>
    </span>
  );
}

export default function TestResultItem({ assertion, index }: TestResultItemProps) {
  const showDetails = !assertion.passed && (assertion.actual || assertion.expected);

  return (
    <div
      className={`flex items-start gap-3 px-4 py-3 border-b border-yapi-border/20 transition-colors ${
        assertion.passed
          ? "hover:bg-yapi-success/5"
          : "hover:bg-yapi-error/5 bg-yapi-error/5"
      }`}
      style={{
        animation: `fadeSlideIn 0.2s ease-out ${index * 0.05}s both`,
      }}
    >
      <StatusIcon passed={assertion.passed} />
      <div className="flex-1 min-w-0">
        <span className="text-sm font-mono text-yapi-fg break-all">
          {assertion.expression}
        </span>

        {showDetails && (
          <div className="mt-2 p-3 bg-yapi-error/10 rounded-lg border border-yapi-error/20">
            {assertion.expected && (
              <div className="flex gap-2 text-xs font-mono mb-1">
                <span className="text-yapi-fg-muted w-16 shrink-0">Expected:</span>
                <span className="text-yapi-error break-all">{assertion.expected}</span>
              </div>
            )}
            {assertion.actual && (
              <div className="flex gap-2 text-xs font-mono">
                <span className="text-yapi-fg-muted w-16 shrink-0">Actual:</span>
                <span className="text-yapi-warning break-all">{assertion.actual}</span>
              </div>
            )}
          </div>
        )}

        {assertion.error && (
          <div className="mt-2 p-2 bg-yapi-error/10 rounded text-xs font-mono text-yapi-error">
            {assertion.error}
          </div>
        )}
      </div>
    </div>
  );
}
