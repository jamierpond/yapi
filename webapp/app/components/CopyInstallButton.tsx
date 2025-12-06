'use client';

import { useState, useEffect } from "react";

type Platform = "mac" | "linux" | "windows";

const installCommands: Record<Platform, { cmd: string; display: string }> = {
  mac: {
    cmd: "curl -fsSL https://yapi.run/install/mac.sh | bash",
    display: "curl -fsSL yapi.run/install/mac.sh | bash",
  },
  linux: {
    cmd: "curl -fsSL https://yapi.run/install/linux.sh | bash",
    display: "curl -fsSL yapi.run/install/linux.sh | bash",
  },
  windows: {
    cmd: "irm https://yapi.run/install/windows.ps1 | iex",
    display: "irm yapi.run/install/windows.ps1 | iex",
  },
};

function detectPlatform(): Platform {
  if (typeof window === "undefined") return "mac";
  const ua = navigator.userAgent.toLowerCase();
  if (ua.includes("win")) return "windows";
  if (ua.includes("linux")) return "linux";
  return "mac";
}

export default function CopyInstallButton() {
  const [platform, setPlatform] = useState<Platform>("mac");
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    setPlatform(detectPlatform());
  }, []);

  const copyInstall = () => {
    navigator.clipboard.writeText(installCommands[platform].cmd);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="flex flex-col gap-2 w-full sm:w-auto">
      <div className="flex justify-center gap-1">
        {(["mac", "linux", "windows"] as Platform[]).map((p) => (
          <button
            key={p}
            onClick={() => setPlatform(p)}
            className={`px-3 py-1 text-xs font-mono rounded transition-colors ${
              platform === p
                ? "bg-yapi-accent text-white"
                : "bg-yapi-bg-elevated/50 text-yapi-fg-muted hover:text-yapi-fg"
            }`}
          >
            {p === "mac" ? "macOS" : p === "linux" ? "Linux" : "Windows"}
          </button>
        ))}
      </div>
      <button
        onClick={copyInstall}
        className="group relative px-6 py-4 bg-black/40 border border-yapi-border hover:border-yapi-accent/50 rounded-xl transition-all duration-300 text-left flex items-center gap-4 font-mono text-sm overflow-hidden min-w-[300px]"
      >
        <div className="absolute inset-0 bg-yapi-accent/5 translate-y-full group-hover:translate-y-0 transition-transform duration-300"></div>
        <span className="text-yapi-accent mr-1 z-10 font-bold">{platform === "windows" ? ">" : "$"}</span>
        <span className="text-yapi-fg-muted z-10 flex-1 truncate">{installCommands[platform].display}</span>
        <span className={`text-yapi-fg-subtle group-hover:text-yapi-accent transition-all whitespace-nowrap z-10 ${copied ? 'scale-110 font-bold text-yapi-success' : ''}`}>
          {copied ? "Copied" : "Copy"}
        </span>
      </button>
    </div>
  );
}
