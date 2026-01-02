"use client";

interface WarningsDisplayProps {
  warnings: string[];
}

export default function WarningsDisplay({ warnings }: WarningsDisplayProps) {
  return (
    <div className="h-full overflow-auto p-6">
      <div className="space-y-3">
        {warnings.map((warning, idx) => (
          <div
            key={idx}
            className="flex items-start gap-3 p-4 bg-yapi-warning/5 border border-yapi-warning/20 rounded-lg"
          >
            <span className="text-yapi-warning text-sm flex-shrink-0 mt-0.5">!</span>
            <p className="text-sm text-yapi-fg-muted leading-relaxed flex-1">
              {warning}
            </p>
          </div>
        ))}
      </div>
    </div>
  );
}
