'use client';

import { useState, useEffect, useRef } from 'react';

type OS = 'mac' | 'linux' | 'windows';

const CONFIG: Record<OS, { label: string; cmd: string; icon: React.ReactNode }> = {
  mac: {
    label: 'macOS',
    cmd: 'curl -fsSL https://yapi.run/install/mac.sh | bash',
    icon: <AppleIcon />,
  },
  linux: {
    label: 'Linux',
    cmd: 'curl -fsSL https://yapi.run/install/linux.sh | bash',
    icon: <LinuxIcon />,
  },
  windows: {
    label: 'Windows',
    cmd: 'irm https://yapi.run/install/windows.ps1 | iex',
    icon: <WindowsIcon />,
  },
};

export default function CopyInstallButton() {
  const [os, setOs] = useState<OS>('mac');
  const [isOpen, setIsOpen] = useState(false);
  const [copied, setCopied] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Auto-detect OS
  useEffect(() => {
    const ua = window.navigator.userAgent.toLowerCase();
    if (ua.includes('win')) setOs('windows');
    else if (ua.includes('linux')) setOs('linux');
    else setOs('mac');
  }, []);

  // Close dropdown on outside click
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(CONFIG[os].cmd);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy', err);
    }
  };

  const handleSelectPlatform = (newOs: OS) => {
    setOs(newOs);
    setIsOpen(false);
    setCopied(false);
  };

  return (
    <div className="relative z-50 font-sans" ref={dropdownRef}>
      {/* --- Main Action Group --- */}
      <div className="inline-flex items-center p-1 rounded-xl bg-yapi-bg-elevated border border-yapi-border shadow-xl backdrop-blur-md">

        {/* Copy Button (Primary Action) */}
        <button
          onClick={handleCopy}
          className="flex items-center gap-3 px-5 py-3 rounded-lg bg-yapi-fg text-yapi-bg hover:bg-white transition-all active:scale-[0.98] group"
        >
          <span className="text-lg">
            {CONFIG[os].icon}
          </span>
          <div className="flex flex-col items-start text-sm">
            <span className="font-bold leading-none">
              {copied ? 'Copied!' : 'Install Script'}
            </span>
            <span className="text-[10px] opacity-70 font-mono mt-1 leading-none">
               {CONFIG[os].label}
            </span>
          </div>
          <div className={`ml-1 transition-transform ${copied ? 'scale-110 text-green-600' : 'group-hover:translate-x-0.5'}`}>
             {copied ? <CheckIcon /> : <CopyIcon />}
          </div>
        </button>

        {/* Separator */}
        <div className="w-[1px] h-8 bg-yapi-border/50 mx-1"></div>

        {/* Dropdown Toggle */}
        <button
          onClick={() => setIsOpen(!isOpen)}
          className={`p-3 rounded-lg hover:bg-white/5 transition-colors text-yapi-fg-muted hover:text-yapi-fg ${isOpen ? 'bg-white/10 text-yapi-fg' : ''}`}
          aria-label="Select Platform"
        >
          <ChevronDown className={`transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`} />
        </button>
      </div>

      {/* --- Dropdown Menu --- */}
      {isOpen && (
        <div className="absolute top-full left-0 mt-2 w-72 p-2 rounded-xl border border-yapi-border bg-[#18181b] shadow-2xl animate-in fade-in zoom-in-95 duration-100 origin-top-left">
          <div className="text-[10px] font-bold text-yapi-fg-subtle uppercase tracking-wider px-2 py-1.5 mb-1">
            Select Platform
          </div>

          {(Object.keys(CONFIG) as OS[]).map((key) => (
            <button
              key={key}
              onClick={() => handleSelectPlatform(key)}
              className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-colors
                ${os === key ? 'bg-yapi-accent/10 text-yapi-accent' : 'text-yapi-fg-muted hover:bg-white/5 hover:text-yapi-fg'}
              `}
            >
              <span className="opacity-80">{CONFIG[key].icon}</span>
              <span>{CONFIG[key].label}</span>
              {os === key && <CheckIcon className="ml-auto w-4 h-4" />}
            </button>
          ))}

          {/* Preview of command */}
          <div className="mt-2 pt-2 border-t border-white/5 px-2 pb-1">
             <div className="text-[10px] text-yapi-fg-subtle mb-1.5">Command Preview:</div>
             <div className="font-mono text-[10px] text-yapi-fg-muted/60 break-all leading-relaxed select-all">
                {CONFIG[os].cmd}
             </div>
          </div>
        </div>
      )}
    </div>
  );
}

/* --- Simple Icons --- */

function CheckIcon({ className = "w-5 h-5" }: { className?: string }) {
  return <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}><path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" /></svg>;
}

function CopyIcon() {
  return <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>;
}

function ChevronDown({ className }: { className?: string }) {
  return <svg className={`w-5 h-5 ${className}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" /></svg>;
}

function AppleIcon() { return <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.8-1.31.02-2.3-1.23-3.14-2.47-1.71-2.45-3.03-6.93-1.26-10.03 1.07-1.87 2.92-2.97 4.96-3.02 1.35-.03 2.62.91 3.44.91.82 0 2.36-1.13 3.98-.96 1.34.1 2.54.67 3.32 1.77-2.91 1.76-2.43 6.01.52 7.32-.42 1.28-1 2.87-2.21 4zM15.5 5.25c.7-1.18 1.15-2.52 1.02-3.87-1.4.11-3.09.93-4.09 2.13-.67.79-1.26 2.06-1.1 3.28 1.57.12 3.16-.78 4.17-1.54z"/></svg>; }
function LinuxIcon() { return <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z"/><circle cx="12" cy="12" r="5" /></svg>; }
function WindowsIcon() { return <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><path d="M0 3.449L9.75 2.1v9.451H0V3.449zm10.949-1.551L24 0v11.4H10.949V1.898zM0 12.6h9.75v9.451L0 20.551V12.6zm10.949 0H24v11.4l-13.051-1.898V12.6z"/></svg>; }
