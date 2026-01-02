"use client";

interface ErrorDisplayProps {
  errorType: string;
  error: string;
  details?: unknown;
}

// Helper function to strip ANSI color codes
function stripAnsiCodes(text: string): string {
  // eslint-disable-next-line no-control-regex
  return text.replace(/\x1b\[[0-9;]*m/g, "").replace(/\x1b\[/g, "");
}

export default function ErrorDisplay({ errorType, error, details }: ErrorDisplayProps) {
  return (
    <div className="flex-1 overflow-auto bg-yapi-bg p-6">
      <div className="max-w-3xl mx-auto">
        <div className="relative overflow-hidden rounded-xl bg-gradient-to-br from-yapi-error/10 via-yapi-error/5 to-transparent border border-yapi-error/30 backdrop-blur-sm">
          <div className="absolute top-0 right-0 w-32 h-32 bg-yapi-error/20 rounded-full blur-3xl" />

          <div className="relative p-6">
            <div className="flex items-start gap-4 mb-6">
              <div className="flex-shrink-0 w-10 h-10 rounded-full bg-yapi-error/20 border border-yapi-error/30 flex items-center justify-center">
                <span className="text-yapi-error text-lg">!</span>
              </div>
              <div className="flex-1">
                <h4 className="text-sm font-bold text-yapi-error tracking-wide mb-2">
                  {errorType.replace(/_/g, " ")}
                </h4>
                <p className="text-sm text-yapi-fg leading-relaxed">
                  {stripAnsiCodes(error)}
                </p>
              </div>
            </div>

            {!!details && (
              <div className="border-t border-yapi-error/20 pt-4">
                <div className="flex items-center gap-2 mb-3">
                  <span className="text-xs font-semibold text-yapi-fg-muted uppercase tracking-wider">
                    Details
                  </span>
                  <div className="h-px flex-1 bg-gradient-to-r from-yapi-border/50 to-transparent" />
                </div>
                <div className="p-4 bg-yapi-bg/50 border border-yapi-border/50 rounded-lg">
                  <pre className="text-xs text-yapi-fg-subtle font-mono overflow-x-auto leading-relaxed whitespace-pre-wrap">
                    {typeof details === "string"
                      ? stripAnsiCodes(details)
                      : JSON.stringify(details, null, 2)}
                  </pre>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
