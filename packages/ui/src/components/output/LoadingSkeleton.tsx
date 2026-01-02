"use client";

export default function LoadingSkeleton() {
  return (
    <div className="h-full flex flex-col bg-yapi-bg relative overflow-hidden">
      {/* Animated background gradient */}
      <div className="absolute inset-0 opacity-30">
        <div className="absolute inset-0 bg-gradient-to-br from-yapi-accent/20 via-transparent to-transparent animate-pulse" />
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-96 h-96 bg-yapi-accent/10 rounded-full blur-3xl animate-ping-slow" />
      </div>

      {/* Content */}
      <div className="relative flex-1 flex flex-col items-center justify-center gap-6">
        {/* Spinner */}
        <div className="text-5xl animate-spin">🐑</div>

        {/* Text */}
        <div className="flex flex-col items-center gap-2">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 rounded-full bg-yapi-accent animate-pulse shadow-[0_0_12px_rgba(255,102,0,0.6)]" />
            <span className="text-sm font-semibold text-yapi-fg tracking-wider">yapi</span>
          </div>
          <span className="text-xs text-yapi-fg-muted animate-pulse">Executing request...</span>
        </div>

        {/* Progress dots */}
        <div className="flex items-center gap-1.5">
          <div className="w-1.5 h-1.5 rounded-full bg-yapi-accent animate-bounce-dot" style={{ animationDelay: "0ms" }} />
          <div className="w-1.5 h-1.5 rounded-full bg-yapi-accent animate-bounce-dot" style={{ animationDelay: "150ms" }} />
          <div className="w-1.5 h-1.5 rounded-full bg-yapi-accent animate-bounce-dot" style={{ animationDelay: "300ms" }} />
        </div>
      </div>
    </div>
  );
}
