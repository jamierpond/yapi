'use client';
import Link from "next/link";
import { useState } from "react";

export default function Landing() {
  const [copied, setCopied] = useState(false);

  const copyInstall = () => {
    navigator.clipboard.writeText("go install yapi.run/cli/cmd/yapi@latest");
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="min-h-screen flex flex-col bg-yapi-bg relative overflow-hidden font-sans text-yapi-fg selection:bg-yapi-accent selection:text-white">
      {/* Background Noise & Gradients */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-[-20%] left-[-10%] w-[50rem] h-[50rem] bg-yapi-accent/5 rounded-full blur-[120px] opacity-30"></div>
        <div className="absolute bottom-[-20%] right-[-10%] w-[40rem] h-[40rem] bg-red-500/5 rounded-full blur-[120px] opacity-20"></div>
        <div className="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20 mix-blend-soft-light"></div>
      </div>

      {/* Navbar */}
      <nav className="relative z-50 px-6 py-6 border-b border-yapi-border/30 backdrop-blur-md bg-yapi-bg/50">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="text-3xl">🐑</span>
            <span className="text-xl font-bold tracking-tight font-mono">yapi</span>
          </div>
          <div className="flex gap-6 items-center">
            <a href="https://github.com/jpond/yapi" className="text-sm font-medium text-yapi-fg-muted hover:text-yapi-fg transition-colors">
              GitHub
            </a>
            <Link
              href="/playground"
              className="hidden sm:block px-5 py-2 text-sm font-semibold rounded-lg bg-yapi-bg-elevated border border-yapi-border hover:border-yapi-accent transition-colors"
            >
              Open Playground
            </Link>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <main className="flex-1 relative z-10 flex flex-col items-center justify-center pt-20 pb-32 px-6">

        {/* The "Status Page" Shade */}
        <div className="mb-8 animate-fade-in-up">
          <div className="inline-flex items-center gap-3 px-4 py-2 rounded-full border border-red-900/50 bg-red-950/30 backdrop-blur-sm">
            <div className="flex h-2 w-2 relative">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2 w-2 bg-red-500"></span>
            </div>
            <span className="text-xs font-mono text-red-200">
              <span className="line-through opacity-50">Localhost is down</span>
              <span className="ml-2 font-bold text-white">Localhost is offline-first.</span>
            </span>
          </div>
        </div>

        <div className="max-w-5xl w-full text-center space-y-8">
          <h1 className="text-5xl md:text-7xl font-bold tracking-tight leading-[1.1]">
            Why do you need a login <br className="hidden md:block" />
            <span className="bg-gradient-to-r from-yapi-accent to-orange-400 bg-clip-text text-transparent">
              to test localhost?
            </span>
          </h1>

          <p className="text-xl text-yapi-fg-muted max-w-2xl mx-auto leading-relaxed">
            Stop letting cloud outages block your dev loop. <br/>
            <strong>yapi</strong> is the Go-powered, git-backed API client that works when the internet doesn't.
          </p>

          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-8">
            <button
              onClick={copyInstall}
              className="group relative px-8 py-4 bg-yapi-bg-elevated hover:bg-yapi-bg-subtle border border-yapi-border hover:border-yapi-border-strong rounded-xl transition-all w-full sm:w-auto min-w-[300px] text-left flex items-center justify-between font-mono text-sm"
            >
              <span className="text-yapi-fg-muted">$ <span className="text-yapi-fg">go install yapi...</span></span>
              <span className="text-yapi-fg-subtle group-hover:text-yapi-accent transition-colors">
                {copied ? "✓ Copied" : "Copy"}
              </span>
            </button>
            <Link
              href="/playground"
              className="px-8 py-4 rounded-xl bg-yapi-accent hover:bg-yapi-accent-hover text-white font-bold transition-all shadow-lg hover:shadow-yapi-accent/40 w-full sm:w-auto"
            >
              Try in Browser
            </Link>
          </div>
        </div>

        {/* The Comparison (Shade Section) */}
        <div className="max-w-5xl w-full mx-auto mt-32 grid md:grid-cols-2 gap-12 items-center">
          <div className="space-y-8 order-2 md:order-1">
            <div className="space-y-4">
              <h3 className="text-2xl font-bold text-yapi-fg">The "Enterprise" Way</h3>
              <ul className="space-y-3 text-yapi-fg-muted">
                <li className="flex items-center gap-3">
                  <span className="text-red-500">✕</span>
                  <span>Forced cloud sync for local collections</span>
                </li>
                <li className="flex items-center gap-3">
                  <span className="text-red-500">✕</span>
                  <span>"Service Unavailable" means you stop working</span>
                </li>
                <li className="flex items-center gap-3">
                  <span className="text-red-500">✕</span>
                  <span>500MB RAM usage for a GET request</span>
                </li>
                <li className="flex items-center gap-3">
                  <span className="text-red-500">✕</span>
                  <span>Updates that move buttons for no reason</span>
                </li>
              </ul>
            </div>

            <div className="h-px w-full bg-yapi-border/50 md:hidden"></div>

            <div className="space-y-4">
              <h3 className="text-2xl font-bold text-yapi-fg">The yapi Way</h3>
              <ul className="space-y-3 text-yapi-fg-muted">
                <li className="flex items-center gap-3">
                  <span className="text-yapi-success">✓</span>
                  <span>Just a YAML file. Commit it to Git.</span>
                </li>
                <li className="flex items-center gap-3">
                  <span className="text-yapi-success">✓</span>
                  <span>Zero runtime dependencies. It's just Go.</span>
                </li>
                <li className="flex items-center gap-3">
                  <span className="text-yapi-success">✓</span>
                  <span>Interactive TUI + LSP for your editor</span>
                </li>
                <li className="flex items-center gap-3">
                  <span className="text-yapi-success">✓</span>
                  <span>Localhost stays local. Forever.</span>
                </li>
              </ul>
            </div>
          </div>

          {/* Visual Representation */}
          <div className="order-1 md:order-2 bg-yapi-bg-elevated border border-yapi-border rounded-xl p-6 shadow-2xl relative overflow-hidden group">
            <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-red-500 to-yapi-accent"></div>
            <div className="font-mono text-xs text-yapi-fg-muted mb-4 flex justify-between">
              <span>~/dev/project/api</span>
              <span>-- yapi watch</span>
            </div>
            <pre className="font-mono text-sm leading-relaxed overflow-x-auto">
              <code className="block text-yapi-fg">
                <span className="text-yapi-fg-subtle"># config.yapi.yml</span>
                <br/>
                <span className="text-yapi-accent">url</span>: http://localhost:8080<br/>
                <span className="text-yapi-accent">method</span>: POST<br/>
                <span className="text-yapi-accent">body</span>:<br/>
                {"  "}status: "ready"<br/>
                {"  "}deployment: "local"<br/>
                <br/>
                <span className="text-yapi-fg-subtle"># Output</span>
                <br/>
                <span className="text-yapi-success">200 OK</span> <span className="text-yapi-fg-muted">4ms</span><br/>
                {`{`}
                <br/>
                {"  "}"message": "No login required."<br/>
                {`}`}
              </code>
            </pre>

            {/* "Offline" Badge */}
            <div className="absolute top-4 right-4 px-2 py-1 bg-yapi-success/10 border border-yapi-success/30 rounded text-[10px] text-yapi-success font-bold uppercase tracking-wider">
              100% Offline
            </div>
          </div>
        </div>

        {/* Feature Cards */}
        <div className="max-w-7xl w-full mx-auto grid md:grid-cols-3 gap-6 mt-32">
          <FeatureCard
            icon="⚡"
            title="Go Native"
            desc="Rewritten in Go. A single binary that starts instantly. No JVM, no Electron, no lag."
          />
          <FeatureCard
            icon="📺"
            title="TUI & Watch Mode"
            desc="Keep your hands on the keyboard. Live-reload requests as you edit your YAML configs."
          />
          <FeatureCard
            icon="🧠"
            title="LSP Integration"
            desc="We don't force an editor on you. We provide a Language Server so you can use yours."
          />
        </div>

      </main>

      {/* Footer */}
      <footer className="border-t border-yapi-border/50 bg-yapi-bg-elevated/30 py-12 px-6">
        <div className="max-w-7xl mx-auto flex flex-col md:flex-row justify-between items-center gap-6">
          <div className="text-yapi-fg-muted text-sm font-mono">
            rm -rf postman && go install yapi
          </div>
          <div className="flex gap-6">
            <a href="https://github.com/jpond/yapi" className="text-yapi-fg-subtle hover:text-yapi-accent transition-colors text-sm">Source Code</a>
          </div>
        </div>
      </footer>
    </div>
  );
}

function FeatureCard({ icon, title, desc }: { icon: string, title: string, desc: string }) {
  return (
    <div className="p-8 rounded-2xl bg-yapi-bg-elevated/50 border border-yapi-border backdrop-blur-sm hover:border-yapi-accent/30 transition-all">
      <div className="h-12 w-12 rounded-lg bg-yapi-bg-subtle flex items-center justify-center mb-6 text-2xl shadow-inner">
        {icon}
      </div>
      <h3 className="text-xl font-bold mb-3 font-mono">{title}</h3>
      <p className="text-yapi-fg-muted leading-relaxed text-sm">
        {desc}
      </p>
    </div>
  );
}
