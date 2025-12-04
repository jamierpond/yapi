'use client';

import { useState } from "react";

export default function CopyInstallButton() {
  const [copied, setCopied] = useState(false);

  const copyInstall = () => {
    navigator.clipboard.writeText("go install yapi.run/cli/cmd/yapi@latest");
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <button
      onClick={copyInstall}
      className="group relative px-6 py-4 bg-black/40 border border-yapi-border hover:border-yapi-accent/50 rounded-xl transition-all duration-300 text-left flex items-center gap-4 font-mono text-sm overflow-hidden min-w-[300px]"
    >
      <div className="absolute inset-0 bg-yapi-accent/5 translate-y-full group-hover:translate-y-0 transition-transform duration-300"></div>
      <span className="text-yapi-accent mr-1 z-10 font-bold">$</span>
      <span className="text-yapi-fg-muted z-10 flex-1">go install yapi.run...</span>
      <span className={`text-yapi-fg-subtle group-hover:text-yapi-accent transition-all whitespace-nowrap z-10 ${copied ? 'scale-110 font-bold text-yapi-success' : ''}`}>
        {copied ? "✓ Copied" : "Copy"}
      </span>
    </button>
  );
}
