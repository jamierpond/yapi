'use client';

import { useState, useEffect } from 'react';

type OS = 'mac' | 'linux' | 'windows';

const COMMANDS: Record<OS, string> = {
  mac: 'curl -fsSL https://yapi.run/install/mac.sh | bash',
  linux: 'curl -fsSL https://yapi.run/install/linux.sh | bash',
  windows: 'irm https://yapi.run/install/windows.ps1 | iex',
};

export default function CopyInstallButton() {
  const [activeTab, setActiveTab] = useState<OS>('mac');
  const [copied, setCopied] = useState(false);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    const ua = window.navigator.userAgent.toLowerCase();
    if (ua.includes('win')) setActiveTab('windows');
    else if (ua.includes('linux')) setActiveTab('linux');
    else setActiveTab('mac');
  }, []);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(COMMANDS[activeTab]);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy!', err);
    }
  };

  if (!mounted) {
    return <div className="h-[52px] w-full max-w-md bg-yapi-bg-elevated/20 rounded-lg animate-pulse" />;
  }

  return (
    <div className="w-full max-w-xl">
      <div className="flex items-stretch rounded-lg border border-yapi-border bg-[#0a0a0a] overflow-hidden">
        {/* Tabs */}
        <div className="flex flex-col border-r border-yapi-border/50 bg-yapi-bg-elevated/30">
          <TabButton active={activeTab === 'mac'} onClick={() => setActiveTab('mac')} label="mac" />
          <TabButton active={activeTab === 'linux'} onClick={() => setActiveTab('linux')} label="linux" />
          <TabButton active={activeTab === 'windows'} onClick={() => setActiveTab('windows')} label="win" />
        </div>

        {/* Command display */}
        <div className="flex-1 flex items-center px-4 py-3 min-w-0">
          <span className="text-yapi-accent mr-2 select-none font-mono">
            {activeTab === 'windows' ? '>' : '$'}
          </span>
          <code className="flex-1 font-mono text-sm text-yapi-fg-muted truncate">
            {COMMANDS[activeTab]}
          </code>
        </div>

        {/* Copy button */}
        <button
          onClick={handleCopy}
          className={`px-4 border-l border-yapi-border/50 transition-all font-medium text-sm
            ${copied
              ? 'bg-yapi-success/10 text-yapi-success'
              : 'bg-yapi-bg-elevated/30 text-yapi-fg-muted hover:text-yapi-fg hover:bg-yapi-bg-elevated'
            }
          `}
        >
          {copied ? 'Copied!' : 'Copy'}
        </button>
      </div>
    </div>
  );
}

function TabButton({ active, onClick, label }: { active: boolean; onClick: () => void; label: string }) {
  return (
    <button
      onClick={onClick}
      className={`px-3 py-1.5 text-[10px] font-mono uppercase tracking-wide transition-colors
        ${active
          ? 'bg-yapi-accent/10 text-yapi-accent'
          : 'text-yapi-fg-subtle hover:text-yapi-fg-muted hover:bg-white/5'
        }
      `}
    >
      {label}
    </button>
  );
}
