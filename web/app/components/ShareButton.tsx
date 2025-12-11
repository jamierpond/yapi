'use client';

import { useState } from "react";

export default function ShareButton() {
  const [copyStatus, setCopyStatus] = useState<"idle" | "copied">("idle");

  async function handleShare() {
    try {
      await navigator.clipboard.writeText(window.location.href);
      setCopyStatus("copied");
      setTimeout(() => setCopyStatus("idle"), 2000);
    } catch (error) {
      console.error("Failed to copy URL:", error);
    }
  }

  return (
    <button
      onClick={handleShare}
      className="group relative px-4 py-1.5 text-sm font-semibold text-white rounded-lg overflow-hidden transition-all duration-300 hover:scale-105 active:scale-95"
    >
      <div className="absolute inset-0 bg-gradient-to-r from-yapi-accent via-orange-500 to-yapi-accent bg-[length:200%_auto] animate-gradient-shift"></div>
      <span className="relative flex items-center gap-2">
        {copyStatus === "copied" ? "Copied!" : "Share"}
      </span>
      <style>{`
        @keyframes gradient-shift {
          0% { background-position: 0% 50%; }
          50% { background-position: 100% 50%; }
          100% { background-position: 0% 50%; }
        }
        .animate-gradient-shift {
          animation: gradient-shift 3s ease infinite;
        }
      `}</style>
    </button>
  );
}
