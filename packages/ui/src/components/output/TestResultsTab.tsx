"use client";

import type { Assertions, AssertionResult } from "../../types/api-contract";
import TestResultItem from "./TestResultItem";

interface TestResultsTabProps {
  assertions: Assertions;
}

export default function TestResultsTab({ assertions }: TestResultsTabProps) {
  const results = assertions.results || [];

  return (
    <div className="h-full flex flex-col overflow-hidden">
      {/* Results List */}
      <div className="flex-1 overflow-auto">
        {results.length > 0 ? (
          results.map((assertion: AssertionResult, index: number) => (
            <TestResultItem
              key={`${assertion.expression}-${index}`}
              assertion={assertion}
              index={index}
            />
          ))
        ) : (
          <div className="p-4 text-center text-yapi-fg-muted text-sm">
            <p>
              {assertions.passed}/{assertions.total} checks passed
            </p>
            <p className="text-xs mt-1 opacity-70">
              Detailed results not available
            </p>
          </div>
        )}
      </div>

      {/* CSS Animations */}
      <style>{`
        @keyframes fadeSlideIn {
          from {
            opacity: 0;
            transform: translateY(-8px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
      `}</style>
    </div>
  );
}
