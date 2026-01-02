"use client";

interface ResponseHeaderProps {
  statusCode?: number;
  timing: number;
}

function getStatusClass(statusCode: number): string {
  if (statusCode >= 200 && statusCode < 300) {
    return "bg-yapi-success/10 text-yapi-success border border-yapi-success/30";
  }
  if (statusCode >= 400) {
    return "bg-yapi-error/10 text-yapi-error border border-yapi-error/30";
  }
  return "bg-yapi-warning/10 text-yapi-warning border border-yapi-warning/30";
}

export default function ResponseHeader({ statusCode, timing }: ResponseHeaderProps) {
  return (
    <div className="relative flex items-center justify-between px-6 h-16 border-b border-yapi-border/50 bg-yapi-bg-elevated/50 backdrop-blur-sm">
      <div className="absolute inset-0 bg-gradient-to-r from-yapi-accent/5 via-transparent to-transparent opacity-50" />

      <div className="relative flex items-center gap-2">
        <div className="w-1.5 h-1.5 rounded-full bg-yapi-accent shadow-[0_0_8px_rgba(255,102,0,0.5)] animate-pulse" />
        <h3 className="text-xs font-semibold text-yapi-fg tracking-wider">YAPI</h3>
      </div>

      <div className="relative flex items-center gap-4">
        {statusCode !== undefined && statusCode > 0 && (
          <span className={`text-xs font-mono font-semibold px-3 py-1.5 rounded-lg backdrop-blur-sm ${getStatusClass(statusCode)}`}>
            {statusCode}
          </span>
        )}
        <div className="flex items-center gap-2 px-3 py-1.5 bg-yapi-bg-elevated/70 border border-yapi-border/60 rounded-lg backdrop-blur-sm">
          <div className="w-1 h-1 rounded-full bg-yapi-accent animate-pulse" />
          <span className="text-xs text-yapi-fg-muted font-mono font-medium">{timing}ms</span>
        </div>
      </div>
    </div>
  );
}
