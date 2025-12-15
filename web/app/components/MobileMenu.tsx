"use client";

import { useState, useEffect } from "react";
import { createPortal } from "react-dom";
import Link from "next/link";

export default function MobileMenu() {
  const [isOpen, setIsOpen] = useState(false);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const menuContent = isOpen && mounted ? (
        <div
          className="fixed inset-0 z-[9999] md:hidden"
          onClick={() => setIsOpen(false)}
        >
          <div className="absolute inset-0 bg-black/60" />

          <div
            className="absolute top-0 right-0 h-full w-72 bg-yapi-bg border-l border-yapi-border shadow-2xl"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex flex-col h-full">
              <div className="flex items-center justify-between px-6 py-5 border-b border-yapi-border">
                <span className="text-xl font-bold text-yapi-fg">Menu</span>
                <button
                  onClick={() => setIsOpen(false)}
                  className="flex items-center justify-center w-9 h-9 rounded-lg hover:bg-yapi-bg-elevated border border-transparent hover:border-yapi-border transition-all"
                  aria-label="Close menu"
                >
                  <svg
                    className="w-5 h-5 text-yapi-fg"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </button>
              </div>

              <nav className="flex-1 px-4 py-6">
                <div className="flex flex-col gap-2">
                  <Link
                    href="/docs"
                    onClick={() => setIsOpen(false)}
                    className="px-4 py-3 text-base font-semibold text-yapi-fg rounded-lg hover:bg-yapi-bg-elevated border border-transparent hover:border-yapi-accent/50 hover:shadow-[0_0_15px_rgba(255,102,0,0.2)] transition-all duration-300"
                  >
                    Docs
                  </Link>
                  <Link
                    href="/blog"
                    onClick={() => setIsOpen(false)}
                    className="px-4 py-3 text-base font-semibold text-yapi-fg rounded-lg hover:bg-yapi-bg-elevated border border-transparent hover:border-yapi-accent/50 hover:shadow-[0_0_15px_rgba(255,102,0,0.2)] transition-all duration-300"
                  >
                    Blog
                  </Link>
                  <Link
                    href="/playground"
                    onClick={() => setIsOpen(false)}
                    className="px-4 py-3 text-base font-semibold text-yapi-fg rounded-lg hover:bg-yapi-bg-elevated border border-transparent hover:border-yapi-accent/50 hover:shadow-[0_0_15px_rgba(255,102,0,0.2)] transition-all duration-300"
                  >
                    Playground
                  </Link>
                </div>
              </nav>

              <div className="px-4 pb-6 border-t border-yapi-border pt-4">
                <a
                  href="https://github.com/jamierpond/yapi"
                  className="flex items-center justify-center gap-2 px-4 py-3 text-base font-semibold text-yapi-fg rounded-lg bg-yapi-bg-elevated border border-yapi-border hover:border-yapi-accent/50 transition-all"
                >
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                    <path fillRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clipRule="evenodd" />
                  </svg>
                  <span>GitHub</span>
                </a>
              </div>
            </div>
          </div>
        </div>
  ) : null;

  return (
    <>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="md:hidden flex flex-col justify-center items-center w-10 h-10 rounded-lg bg-yapi-bg-elevated border border-yapi-border hover:border-yapi-accent/50 transition-all"
        aria-label="Toggle menu"
      >
        <div className="flex flex-col gap-1.5 w-5">
          <span
            className={`h-0.5 w-full bg-yapi-fg transition-all duration-300 ${
              isOpen ? "rotate-45 translate-y-2" : ""
            }`}
          />
          <span
            className={`h-0.5 w-full bg-yapi-fg transition-all duration-300 ${
              isOpen ? "opacity-0" : ""
            }`}
          />
          <span
            className={`h-0.5 w-full bg-yapi-fg transition-all duration-300 ${
              isOpen ? "-rotate-45 -translate-y-2" : ""
            }`}
          />
        </div>
      </button>
      {mounted && menuContent ? createPortal(menuContent, document.body) : null}
    </>
  );
}
