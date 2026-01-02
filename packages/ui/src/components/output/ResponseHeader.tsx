"use client";

import { useState, useCallback } from "react";

interface ResponseHeaderProps {
  statusCode?: number;
  timing: number;
  onCopy?: () => string | undefined;
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

export default function ResponseHeader({ statusCode, timing, onCopy }: ResponseHeaderProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    if (!onCopy) return;
    const content = onCopy();
    if (!content) return;

    try {
      await navigator.clipboard.writeText(content);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  }, [onCopy]);

  return (
    <div className="relative flex items-center justify-between px-6 h-16 border-b border-yapi-border/50 bg-yapi-bg-elevated/50 backdrop-blur-sm">
      <div className="absolute inset-0 bg-gradient-to-r from-yapi-accent/5 via-transparent to-transparent opacity-50" />

      <div className="relative flex items-center gap-2">
        <div className="w-1.5 h-1.5 rounded-full bg-yapi-accent shadow-[0_0_8px_rgba(255,102,0,0.5)] animate-pulse" />
        <h3 className="text-xs font-semibold text-yapi-fg tracking-wider">YAPI</h3>
      </div>

      <div className="relative flex items-center gap-3">
        {onCopy && (
          <button
            onClick={handleCopy}
            className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg border backdrop-blur-sm transition-all ${
              copied
                ? "bg-yapi-success/10 text-yapi-success border-yapi-success/30"
                : "bg-yapi-bg-elevated/70 text-yapi-fg-muted border-yapi-border/60 hover:text-yapi-fg hover:border-yapi-border"
            }`}
          >
            {copied ? (
              <>
                <CheckIcon />
                Copied
              </>
            ) : (
              <>
                <CopyIcon />
                Copy
              </>
            )}
          </button>
        )}
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

function CopyIcon() {
  return (
    <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
    </svg>
  );
}

function CheckIcon() {
  return (
    <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
    </svg>
  );
}
