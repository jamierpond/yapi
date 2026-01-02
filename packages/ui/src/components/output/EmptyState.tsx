"use client";

export default function EmptyState() {
  return (
    <div className="h-full flex items-center justify-center bg-yapi-bg relative overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-br from-yapi-accent/10 via-transparent to-transparent opacity-60" />
      <div className="relative text-center space-y-4">
        <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-yapi-bg-elevated border border-yapi-border shadow-lg shadow-yapi-accent/10">
          <div className="text-2xl opacity-70">*</div>
        </div>
        <div className="space-y-2">
          <p className="text-sm text-yapi-fg-muted font-medium">Ready to execute</p>
          <p className="text-xs text-yapi-fg-subtle">
            Press{" "}
            <kbd className="px-2 py-1 text-[10px] bg-yapi-bg-elevated border border-yapi-border rounded font-mono">
              Cmd+Enter
            </kbd>{" "}
            to run
          </p>
        </div>
      </div>
    </div>
  );
}
