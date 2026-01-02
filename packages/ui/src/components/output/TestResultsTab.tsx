"use client";

import type { Assertions } from "../../types/api-contract";
import TestSummaryBanner from "./TestSummaryBanner";
import TestResultItem from "./TestResultItem";

interface TestResultsTabProps {
  assertions: Assertions;
}

export default function TestResultsTab({ assertions }: TestResultsTabProps) {
  const results = assertions.results || [];

  return (
    <div className="h-full flex flex-col overflow-hidden">
      {/* Summary Banner */}
      <TestSummaryBanner assertions={assertions} />

      {/* Results List */}
      <div className="flex-1 overflow-auto">
        {results.length > 0 ? (
          results.map((assertion, index) => (
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

        @keyframes successPulse {
          0% {
            box-shadow: 0 0 0 0 rgba(0, 230, 118, 0.4);
          }
          70% {
            box-shadow: 0 0 0 10px rgba(0, 230, 118, 0);
          }
          100% {
            box-shadow: 0 0 0 0 rgba(0, 230, 118, 0);
          }
        }
      `}</style>
    </div>
  );
}
