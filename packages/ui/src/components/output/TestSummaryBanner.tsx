"use client";

import type { Assertions } from "../../types/api-contract";

interface TestSummaryBannerProps {
  assertions: Assertions;
}

export default function TestSummaryBanner({ assertions }: TestSummaryBannerProps) {
  const { total, passed } = assertions;
  const failed = total - passed;
  const allPassed = failed === 0;

  return (
    <div
      className={`px-4 py-3 flex items-center gap-3 border-b ${
        allPassed
          ? "bg-yapi-success/10 border-yapi-success/30"
          : "bg-yapi-error/10 border-yapi-error/30"
      }`}
    >
      {/* Status Icon */}
      <div
        className={`w-8 h-8 rounded-full flex items-center justify-center text-lg ${
          allPassed
            ? "bg-yapi-success/20 text-yapi-success"
            : "bg-yapi-error/20 text-yapi-error"
        }`}
        style={{
          animation: allPassed ? "successPulse 1s ease-out" : undefined,
        }}
      >
        {allPassed ? "✓" : "✗"}
      </div>

      {/* Status Text */}
      <div className="flex-1">
        <p
          className={`text-sm font-semibold ${
            allPassed ? "text-yapi-success" : "text-yapi-error"
          }`}
        >
          {allPassed
            ? total === 1
              ? "Check Passed"
              : "All Checks Passed"
            : `${failed}/${total} Check${failed !== 1 ? "s" : ""} Failed`}
        </p>
        {!allPassed && (
          <p className="text-xs text-yapi-fg-muted mt-0.5">
            {passed}/{total} passed
          </p>
        )}
      </div>

      {/* Stats */}
      <div className="flex items-center gap-2 text-xs font-mono">
        <span className="px-2 py-1 rounded bg-yapi-success/20 text-yapi-success">
          {passed} passed
        </span>
        {failed > 0 && (
          <span className="px-2 py-1 rounded bg-yapi-error/20 text-yapi-error">
            {failed} failed
          </span>
        )}
      </div>
    </div>
  );
}
