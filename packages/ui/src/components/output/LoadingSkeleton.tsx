"use client";

export default function LoadingSkeleton() {
  return (
    <div className="h-full flex flex-col bg-yapi-bg relative">
      {/* Header skeleton */}
      <div className="relative flex items-center justify-between px-6 h-16 border-b border-yapi-border/50 bg-yapi-bg-elevated/50 backdrop-blur-sm">
        <div className="absolute inset-0 bg-gradient-to-r from-yapi-accent/5 via-transparent to-transparent opacity-50" />
        <div className="relative flex items-center gap-2">
          <div className="w-1.5 h-1.5 rounded-full bg-yapi-border animate-pulse" />
          <div className="h-3 w-16 rounded-full yapi-skeleton" />
        </div>
        <div className="relative flex items-center gap-4">
          <div className="h-7 w-14 rounded-lg yapi-skeleton" />
          <div className="h-7 w-20 rounded-lg yapi-skeleton" />
        </div>
      </div>

      {/* Body skeleton */}
      <div className="flex-1 overflow-hidden bg-yapi-bg relative">
        <div className="absolute left-0 top-0 bottom-0 w-14 border-r border-yapi-border/30 flex flex-col items-center gap-[18px] pt-4 pb-4 bg-yapi-bg-editor">
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="w-4 h-3 rounded yapi-skeleton opacity-40" />
          ))}
        </div>
        <div className="absolute left-14 top-0 right-0 bottom-0 p-4 space-y-[18px] bg-yapi-bg-editor">
          {[10, 9, 8, 7, 6, 9, 5, 4].map((w, i) => (
            <div key={i} className={`h-3 w-${w}/12 rounded yapi-skeleton`} />
          ))}
        </div>
      </div>

      <style>{`
        @keyframes yapi-skeleton-shimmer {
          0% { background-position: -200% 0; }
          100% { background-position: 200% 0; }
        }
        .yapi-skeleton {
          background-image: linear-gradient(90deg, rgba(255,255,255,0.02) 0%, rgba(255,255,255,0.08) 20%, rgba(255,255,255,0.02) 40%);
          background-color: var(--color-yapi-bg-editor);
          background-size: 200% 100%;
          animation: yapi-skeleton-shimmer 1.2s ease-in-out infinite;
        }
      `}</style>
    </div>
  );
}
