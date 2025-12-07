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

  const cyclePlatform = (e: React.MouseEvent) => {
    e.stopPropagation();
    const platforms: Platform[] = ["mac", "linux", "windows"];
    const idx = platforms.indexOf(platform);
    setPlatform(platforms[(idx + 1) % 3]);
  };

  return (
    <button
      onClick={copyInstall}
      className="group relative px-6 py-4 bg-black/40 border border-yapi-border hover:border-yapi-accent/50 rounded-xl transition-all duration-300 text-left flex items-center gap-3 font-mono text-sm overflow-hidden w-full sm:w-auto min-w-[420px]"
    >
      <div className="absolute inset-0 bg-yapi-accent/5 translate-y-full group-hover:translate-y-0 transition-transform duration-300"></div>

      <span
        onClick={cyclePlatform}
        className="z-10 px-2 py-0.5 text-[10px] rounded bg-yapi-bg-elevated/50 text-yapi-fg-subtle hover:text-yapi-accent cursor-pointer transition-colors"
      >
        {platform === "mac" ? "macOS" : platform === "linux" ? "Linux" : "Windows"}
      </span>

      <span className="text-yapi-accent z-10 font-bold">{platform === "windows" ? ">" : "$"}</span>
      <span className="text-yapi-fg-muted z-10 flex-1 text-xs sm:text-sm">
        {installCommands[platform].display}
      </span>
      <span className={`text-yapi-fg-subtle group-hover:text-yapi-accent transition-all whitespace-nowrap z-10 text-xs ${copied ? 'font-bold text-yapi-success' : ''}`}>
        {copied ? "Copied!" : "Copy"}
      </span>
    </button>
  );
}
