"use client";

import { useState, useCallback } from "react";
import type { Assertions } from "../../types/api-contract";

interface ResponseHeaderProps {
  statusCode?: number;
  timing: number;
  onCopy?: () => string | undefined;
  assertions?: Assertions;
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

export default function ResponseHeader({ statusCode, timing, onCopy, assertions }: ResponseHeaderProps) {
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

  const hasTests = assertions && assertions.total > 0;
  const allTestsPassed = hasTests && assertions.passed === assertions.total;

  return (
    <div className="relative flex items-center justify-between px-6 h-16 border-b border-yapi-border/50 bg-yapi-bg-elevated/50 backdrop-blur-sm">
      <div className="absolute inset-0 bg-gradient-to-r from-yapi-accent/5 via-transparent to-transparent opacity-50" />

      <div className="relative flex items-center gap-3">
        <a
          href="https://github.com/jamierpond/yapi"
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center gap-2 hover:opacity-80 transition-opacity"
        >
          <div className="w-1.5 h-1.5 rounded-full bg-yapi-accent shadow-[0_0_8px_rgba(255,102,0,0.5)] animate-pulse" />
          <h3 className="text-xs font-semibold text-yapi-fg tracking-wider">yapi</h3>
        </a>
        <a
          href="https://github.com/jamierpond/yapi/issues/new"
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center gap-1.5 px-2 py-1 text-xs font-medium rounded border bg-yapi-bg-elevated/70 text-yapi-fg-muted border-yapi-border/60 hover:text-yapi-fg hover:border-yapi-border transition-all"
        >
          <IssueIcon />
          Create Issue
        </a>
      </div>

      <div className="relative flex items-center gap-3">
        {/* Test Status Badge */}
        {hasTests && (
          <div
            className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-mono font-medium rounded-lg border backdrop-blur-sm ${
              allTestsPassed
                ? "bg-yapi-success/10 text-yapi-success border-yapi-success/30"
                : "bg-yapi-error/10 text-yapi-error border-yapi-error/30"
            }`}
          >
            {allTestsPassed ? <TestPassIcon /> : <TestFailIcon />}
            {assertions.passed}/{assertions.total}
          </div>
        )}
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

function IssueIcon() {
  return (
    <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 16 16">
      <path d="M8 9.5a1.5 1.5 0 1 0 0-3 1.5 1.5 0 0 0 0 3Z" />
      <path d="M8 0a8 8 0 1 1 0 16A8 8 0 0 1 8 0ZM1.5 8a6.5 6.5 0 1 0 13 0 6.5 6.5 0 0 0-13 0Z" />
    </svg>
  );
}

function TestPassIcon() {
  return (
    <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M5 13l4 4L19 7" />
    </svg>
  );
}

function TestFailIcon() {
  return (
    <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M6 18L18 6M6 6l12 12" />
    </svg>
  );
}
