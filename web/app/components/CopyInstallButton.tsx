'use client';

import { useState, useEffect } from 'react';

type OS = 'mac' | 'linux' | 'windows';

const OS_NAMES: Record<OS, string> = {
  mac: 'macOS',
  linux: 'Linux',
  windows: 'Windows',
};

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
    const platform = window.navigator.platform.toLowerCase();
    if (platform.startsWith('mac')) {
      setActiveTab('mac');
    } else if (platform.startsWith('win')) {
      setActiveTab('windows');
    } else if (platform.includes('linux')) {
      setActiveTab('linux');
    }
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
    return <div className="h-[108px] w-full max-w-xl bg-yapi-bg-elevated/20 rounded-lg animate-pulse" />;
  }

  return (
    <div className="w-full max-w-xl">
      <div className="flex items-center justify-center gap-2 mb-2">
        {(['mac', 'linux', 'windows'] as OS[]).map((os) => (
          <button
            key={os}
            onClick={() => setActiveTab(os)}
            className={`px-3 py-1 text-sm font-medium rounded-md transition-colors ${
              activeTab === os
                ? 'bg-yapi-accent/10 text-yapi-accent'
                : 'text-yapi-fg-muted hover:bg-yapi-bg-elevated'
            }`}
          >
            {OS_NAMES[os]}
          </button>
        ))}
      </div>
      <div className="relative flex items-center rounded-lg border border-yapi-border bg-[#0a0a0a] overflow-hidden">
        {/* Command display */}
        <div className="flex-1 flex items-center bg-yapi-bg-elevated/10 px-4 py-3 min-w-0 h-[48px]">
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
          className={`px-4 border-l border-yapi-border/50 transition-colors font-medium text-sm flex items-center justify-center w-24 h-[48px] flex-shrink-0
            ${copied
              ? 'bg-yapi-success/10 text-yapi-success'
              : 'bg-yapi-bg-elevated/30 text-yapi-fg-muted hover:text-yapi-fg hover:bg-yapi-bg-elevated'
            }
          `}
        >
          {copied ? (
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="w-5 h-5">
              <path fillRule="evenodd" d="M16.704 4.153a.75.75 0 01.143 1.052l-8 10.5a.75.75 0 01-1.127.075l-4.5-4.5a.75.75 0 011.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 011.052-.143z" clipRule="evenodd" />
            </svg>
          ) : (
            'Copy'
          )}
        </button>
      </div>
    </div>
  );
}

