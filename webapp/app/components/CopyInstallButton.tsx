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

  // Auto-detect OS on mount
  useEffect(() => {
    setMounted(true);
    const userAgent = window.navigator.userAgent.toLowerCase();
    if (userAgent.includes('win')) {
      setActiveTab('windows');
    } else if (userAgent.includes('linux')) {
      setActiveTab('linux');
    } else {
      setActiveTab('mac');
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

  // Prevent hydration mismatch by rendering a skeleton until mounted
  if (!mounted) return <div className="h-[140px] w-full max-w-lg bg-yapi-bg-elevated/20 rounded-xl animate-pulse" />;

  return (
    <div className="w-full max-w-lg mx-auto">
      <div className="relative group rounded-xl border border-yapi-border bg-[#121212] shadow-2xl overflow-hidden">

        {/* Glowing border effect on hover */}
        <div className="absolute -inset-0.5 bg-gradient-to-r from-yapi-accent/30 to-purple-500/30 rounded-xl blur opacity-0 group-hover:opacity-100 transition duration-500 group-hover:duration-200" />

        <div className="relative bg-[#121212] rounded-xl flex flex-col">

          {/* Header / Tabs */}
          <div className="flex items-center border-b border-white/5 bg-white/5 px-2 pt-2">
            <TabButton
              active={activeTab === 'mac'}
              onClick={() => setActiveTab('mac')}
              icon={<AppleIcon />}
              label="macOS"
            />
            <TabButton
              active={activeTab === 'linux'}
              onClick={() => setActiveTab('linux')}
              icon={<LinuxIcon />}
              label="Linux"
            />
            <TabButton
              active={activeTab === 'windows'}
              onClick={() => setActiveTab('windows')}
              icon={<WindowsIcon />}
              label="Windows"
            />
          </div>

          {/* Code Area */}
          <div className="p-4 flex items-center justify-between gap-4">
            <div className="flex-1 font-mono text-sm overflow-x-auto whitespace-nowrap scrollbar-hide text-yapi-fg-muted selection:bg-yapi-accent selection:text-white">
              <span className="text-yapi-accent mr-2 select-none">{activeTab === 'windows' ? '>' : '$'}</span>
              {COMMANDS[activeTab]}
            </div>

            <button
              onClick={handleCopy}
              className={`
                flex-shrink-0 flex items-center justify-center h-10 w-10 rounded-md transition-all duration-200 border
                ${copied
                  ? 'bg-yapi-success/10 border-yapi-success text-yapi-success'
                  : 'bg-yapi-bg-elevated hover:bg-white/10 border-white/10 text-yapi-fg hover:text-white'
                }
              `}
              aria-label="Copy to clipboard"
            >
              {copied ? <CheckIcon /> : <CopyIcon />}
            </button>
          </div>
        </div>
      </div>

      {/* Helper text below */}
      <div className="mt-3 text-center">
        <p className="text-xs text-yapi-fg-subtle opacity-60 font-mono">
          Paste this into your terminal to install <span className="text-yapi-accent">yapi</span>
        </p>
      </div>
    </div>
  );
}

function TabButton({ active, onClick, icon, label }: { active: boolean; onClick: () => void; icon: React.ReactNode; label: string }) {
  return (
    <button
      onClick={onClick}
      className={`
        flex items-center gap-2 px-4 py-2 text-xs font-medium rounded-t-lg transition-colors relative top-[1px]
        ${active
          ? 'bg-[#121212] text-yapi-accent border-t border-l border-r border-white/5'
          : 'text-yapi-fg-muted hover:text-yapi-fg hover:bg-white/5'
        }
      `}
    >
      {icon}
      {label}
    </button>
  );
}

/* --- Icons --- */

function CheckIcon() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="20 6 9 17 4 12"></polyline>
    </svg>
  );
}

function CopyIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
      <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
    </svg>
  );
}

function AppleIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" className="opacity-80">
      <path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z"/>
    </svg>
  );
}

function LinuxIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" className="opacity-80">
      <path d="M12.504 0c-.155 0-.315.008-.48.021-4.226.333-3.105 4.807-3.17 6.298-.076 1.092-.3 1.953-1.05 3.02-.885 1.051-2.127 2.75-2.716 4.521-.278.832-.41 1.684-.287 2.489a.424.424 0 00-.11.135c-.26.268-.45.6-.663.839-.199.199-.485.267-.797.4-.313.136-.658.269-.864.68-.09.189-.136.394-.132.602 0 .199.027.4.055.536.058.399.116.728.04.97-.249.68-.28 1.145-.106 1.484.174.334.535.47.94.601.81.2 1.91.135 2.774.6.926.466 1.866.67 2.616.47.526-.116.97-.464 1.208-.946.587-.003 1.23-.269 2.26-.334.699-.058 1.574.267 2.577.2.025.134.063.198.114.333l.003.003c.391.778 1.113 1.132 1.884 1.071.771-.06 1.592-.536 2.257-1.306.631-.765 1.683-1.084 2.378-1.503.348-.199.629-.469.649-.853.023-.4-.2-.811-.714-1.376v-.097l-.003-.003c-.17-.2-.25-.535-.338-.926-.085-.401-.182-.786-.492-1.046h-.003c-.059-.054-.123-.067-.188-.135a.357.357 0 00-.19-.064c.431-1.278.264-2.55-.173-3.694-.533-1.41-1.465-2.638-2.175-3.483-.796-1.005-1.576-1.957-1.56-3.368.026-2.152.236-6.133-3.544-6.139zm.529 3.405h.013c.213 0 .396.062.584.198.19.135.33.332.438.533.105.259.158.459.166.724 0-.02.006-.04.006-.06v.105a.086.086 0 01-.004-.021l-.004-.024a1.807 1.807 0 01-.15.706.953.953 0 01-.213.335.71.71 0 00-.088-.042c-.104-.045-.198-.064-.284-.133a1.312 1.312 0 00-.22-.066c.05-.06.146-.133.183-.198.053-.128.082-.264.088-.402v-.02a1.21 1.21 0 00-.061-.4c-.045-.134-.101-.2-.183-.333-.084-.066-.167-.132-.267-.132h-.016c-.093 0-.176.03-.262.132a.8.8 0 00-.205.334 1.18 1.18 0 00-.09.4v.019c.002.089.008.179.02.267-.193-.067-.438-.135-.607-.202a1.635 1.635 0 01-.018-.2v-.02a1.772 1.772 0 01.15-.768c.082-.22.232-.406.43-.533a.985.985 0 01.594-.2zm-2.962.059h.036c.142 0 .27.048.399.135.146.129.264.288.344.465.09.199.14.4.153.667v.004c.007.134.006.2-.002.266v.08c-.03.007-.056.018-.083.024-.152.055-.274.135-.393.2.012-.09.013-.18.003-.267v-.015c-.012-.133-.04-.2-.082-.333a.613.613 0 00-.166-.267.248.248 0 00-.183-.064h-.021c-.071.006-.13.04-.186.132a.552.552 0 00-.12.27.944.944 0 00-.023.33v.015c.012.135.037.2.08.334.046.134.098.2.166.268.01.009.02.018.034.024-.07.057-.117.07-.176.136a.304.304 0 01-.131.068 2.62 2.62 0 01-.275-.402 1.772 1.772 0 01-.155-.667 1.759 1.759 0 01.08-.668 1.43 1.43 0 01.283-.535c.128-.133.26-.2.418-.2zm1.37 1.706c.332 0 .733.065 1.216.399.293.2.523.269 1.052.468h.003c.255.136.405.266.478.399v-.131a.571.571 0 01.016.47c-.123.31-.516.643-1.063.842v.002c-.268.135-.501.333-.775.465-.276.135-.588.292-1.012.267a1.139 1.139 0 01-.448-.067 3.566 3.566 0 01-.322-.198c-.195-.135-.363-.332-.612-.465v-.005h-.005c-.4-.246-.616-.512-.686-.71-.07-.268-.005-.47.193-.6.224-.135.38-.271.483-.336.104-.074.143-.102.176-.131h.002v-.003c.169-.202.436-.47.839-.601.139-.036.294-.065.466-.065zm2.8 2.142c.358 1.417 1.196 3.475 1.735 4.473.286.534.855 1.659 1.102 3.024.156-.005.33.018.513.064.646-1.671-.546-3.467-1.089-3.966-.22-.2-.232-.335-.123-.335.59.534 1.365 1.572 1.646 2.757.13.535.16 1.104.021 1.67.067.028.135.06.205.067 1.032.534 1.413.938 1.23 1.537v-.043c-.06-.003-.12 0-.18 0h-.016c.151-.467-.182-.825-1.065-1.224-.915-.4-1.646-.336-1.77.465-.008.043-.013.066-.018.135-.068.023-.139.053-.209.064-.43.268-.662.669-.793 1.187-.13.533-.17 1.156-.205 1.869v.003c-.02.482-.04.965-.07 1.39a.96.96 0 01-.109.469c-.152.37-.444.667-.755.867-.358.268-.596.465-.596.868 0 .4.18.735.49.935.16.1.32.167.457.233.06.025.12.05.18.076.53.268.918.201 1.142-.132.091-.135.15-.242.2-.333.084-.153.137-.27.208-.398.135-.2.27-.398.539-.398.203 0 .4.134.534.331v-.003c.07.066.137.133.2.267.134.2.2.334.271.467.07.134.151.268.212.402.123.27.212.602.054.87-.066.131-.143.267-.32.4-.179.135-.406.202-.703.202l.003-.002-.003.002h-.003c-.4.02-.706-.132-.933-.334-.227-.2-.399-.469-.532-.667-.134-.2-.227-.401-.291-.535a.46.46 0 00-.08-.135c-.065.066-.097.2-.13.333-.062.198-.088.532-.166.866-.109.468-.356.864-.755 1.005-.4.134-.869.066-1.402-.202-.531-.266-1.067-.467-1.535-.467-.468 0-.867.2-1.167.599-.065.066-.135.066-.202.066-.065 0-.133-.024-.2-.067h-.001a.96.96 0 01-.233-.268l-.002-.001c-.155-.066-.355-.198-.589-.465-.228-.267-.483-.666-.69-1.2-.206-.533-.354-1.198-.354-1.998 0-.799.146-1.532.41-2.133.132-.3.293-.566.49-.8.196-.232.436-.434.742-.601.306-.165.66-.268 1.09-.335.42-.065.902-.1 1.457-.1.555 0 1.124.035 1.7.1.576.067 1.157.167 1.735.3.58.135 1.15.302 1.697.5.55.2 1.066.434 1.536.7.134.066.227.133.287.2-.13-.532-.268-.867-.473-1.133a.857.857 0 00-.343-.27c-.2-.131-.543-.267-.88-.398-.34-.134-.69-.268-.992-.4-.3-.133-.556-.266-.762-.467a.97.97 0 01-.336-.534.948.948 0 01.01-.534c.047-.134.124-.266.243-.4l.085-.085v-.007c.283-.333.87-.735 1.27-.935l.132-.066c-.01-.067-.02-.133-.036-.2-.17-.866-.934-1.334-1.636-1.666l-.003-.003c-.3-.133-.467-.267-.533-.4a.637.637 0 01-.053-.467c.093-.4.43-.668.775-.867a.93.93 0 01.177-.068c.005-.068.008-.135.014-.2.065-.932.321-1.998.868-2.932.547-.932 1.342-1.73 2.459-2.198.111-.05.227-.092.345-.133z"/>
    </svg>
  );
}

function WindowsIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" className="opacity-80">
      <path d="M0 3.449L9.75 2.1v9.451H0V3.449zm10.949-1.551L24 0v11.4H10.949V1.898zM0 12.6h9.75v9.451L0 20.551V12.6zm10.949 0H24v11.4l-13.051-1.898V12.6z"/>
    </svg>
  );
}
