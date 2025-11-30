import Link from "next/link";

export default function Landing() {
  return (
    <div className="min-h-screen flex flex-col bg-yapi-bg relative overflow-hidden">
      {/* Background accents */}
      <div className="absolute inset-0 bg-gradient-to-br from-yapi-accent/10 via-transparent to-transparent opacity-60 pointer-events-none"></div>
      <div className="absolute top-1/4 right-1/4 w-96 h-96 bg-yapi-accent/5 rounded-full blur-3xl pointer-events-none"></div>

      {/* Header */}
      <header className="relative px-6 py-6 border-b border-yapi-border/50 bg-yapi-bg-elevated/50 backdrop-blur-sm">
        <div className="max-w-6xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="text-2xl">🐑</div>
            <h1 className="text-lg font-bold text-yapi-fg font-mono">yapi</h1>
          </div>
          <Link
            href="/playground"
            className="px-4 py-2 text-sm font-semibold rounded-lg bg-yapi-bg-elevated border border-yapi-border hover:border-yapi-accent/50 text-yapi-fg transition-all duration-300"
          >
            Playground
          </Link>
        </div>
      </header>

      {/* Hero Section */}
      <main className="flex-1 relative">
        <div className="max-w-4xl mx-auto px-6 py-20">
          {/* Hero */}
          <div className="text-center space-y-8 mb-20">
            <div className="inline-flex items-center gap-2 px-4 py-2 bg-yapi-bg-elevated border border-yapi-border rounded-full">
              <div className="w-2 h-2 rounded-full bg-yapi-accent animate-pulse"></div>
              <span className="text-xs font-mono text-yapi-fg-muted uppercase tracking-wider">
                Bash-powered YAML API workbench
              </span>
            </div>

            <h2 className="text-5xl md:text-6xl font-bold text-yapi-fg font-mono leading-tight">
              API testing
              <br />
              <span className="text-yapi-accent">in YAML</span>
            </h2>

            <p className="text-xl text-yapi-fg-muted max-w-2xl mx-auto leading-relaxed">
              A small, Bash-powered client that speaks HTTP, gRPC, and raw TCP.
              Write clean YAML configs, version control your requests, execute from CLI or web.
            </p>

            <div className="flex items-center justify-center gap-4 pt-4">
              <Link
                href="/playground"
                className="group relative px-8 py-4 text-lg font-semibold rounded-lg transition-all duration-300 flex items-center gap-3 overflow-hidden bg-gradient-to-r from-yapi-accent to-yapi-accent hover:from-yapi-accent hover:to-orange-500 text-white shadow-lg hover:shadow-xl hover:shadow-yapi-accent/30 hover:scale-105 active:scale-95"
              >
                <div className="absolute inset-0 bg-gradient-to-r from-white/0 via-white/20 to-white/0 opacity-0 group-hover:opacity-100 transition-opacity duration-500 rounded-lg animate-shimmer"></div>
                <span className="relative">Try Playground</span>
                <span className="relative text-2xl">→</span>
              </Link>

              <a
                href="https://github.com/jamierpond/yapi"
                className="px-8 py-4 text-lg font-semibold rounded-lg bg-yapi-bg-elevated border border-yapi-border hover:border-yapi-accent/50 text-yapi-fg transition-all duration-300"
              >
                View on GitHub
              </a>
            </div>
          </div>

          {/* Features */}
          <div className="grid md:grid-cols-3 gap-6 mb-20">
            <div className="p-6 bg-yapi-bg-elevated border border-yapi-border/50 rounded-xl hover:border-yapi-accent/30 transition-colors duration-300">
              <div className="text-3xl mb-4">⚡</div>
              <h3 className="text-lg font-bold text-yapi-fg mb-2 font-mono">
                Simple YAML
              </h3>
              <p className="text-sm text-yapi-fg-muted leading-relaxed">
                Define requests in clean YAML. No complex SDKs, no verbose code. Just readable configs you can version control.
              </p>
            </div>

            <div className="p-6 bg-yapi-bg-elevated border border-yapi-border/50 rounded-xl hover:border-yapi-accent/30 transition-colors duration-300">
              <div className="text-3xl mb-4">🔌</div>
              <h3 className="text-lg font-bold text-yapi-fg mb-2 font-mono">
                Multi-Protocol
              </h3>
              <p className="text-sm text-yapi-fg-muted leading-relaxed">
                HTTP, gRPC, and raw TCP in one tool. Switch protocols with a single field change.
              </p>
            </div>

            <div className="p-6 bg-yapi-bg-elevated border border-yapi-border/50 rounded-xl hover:border-yapi-accent/30 transition-colors duration-300">
              <div className="text-3xl mb-4">🛠️</div>
              <h3 className="text-lg font-bold text-yapi-fg mb-2 font-mono">
                Bash Native
              </h3>
              <p className="text-sm text-yapi-fg-muted leading-relaxed">
                Built on Bash. No runtime dependencies. Run anywhere Bash runs. Pipe, redirect, script as needed.
              </p>
            </div>
          </div>

          {/* Quickstart */}
          <div className="mb-20 space-y-8">
            <h3 className="text-2xl font-bold text-yapi-fg mb-6 text-center font-mono">
              Quick Start
            </h3>

            {/* Step 1: Install */}
            <div className="space-y-3">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-yapi-accent/20 border border-yapi-accent/40 flex items-center justify-center text-sm font-bold text-yapi-accent font-mono">
                  1
                </div>
                <h4 className="text-lg font-bold text-yapi-fg font-mono">Install yapi</h4>
              </div>
              <div className="ml-11 bg-yapi-bg-elevated border border-yapi-border/50 rounded-xl p-4">
                <pre className="text-sm text-yapi-fg font-mono overflow-x-auto">
{`# Clone or install to ~/.config/yapi
git clone https://github.com/jamierpond/yapi.git ~/.config/yapi

# Add to your shell (zsh example)
echo 'source ~/.config/yapi/bin/yapi.zsh' >> ~/.zshrc
source ~/.zshrc`}
                </pre>
              </div>
            </div>

            {/* Step 2: Create YAML */}
            <div className="space-y-3">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-yapi-accent/20 border border-yapi-accent/40 flex items-center justify-center text-sm font-bold text-yapi-accent font-mono">
                  2
                </div>
                <h4 className="text-lg font-bold text-yapi-fg font-mono">Create a YAML request file</h4>
              </div>
              <div className="ml-11 bg-yapi-bg-elevated border border-yapi-border/50 rounded-xl p-4">
                <pre className="text-sm text-yapi-fg font-mono leading-relaxed overflow-x-auto">
{`# request.yaml
url: https://api.github.com/users/octocat
method: GET
headers:
  Accept: application/json`}
                </pre>
              </div>
            </div>

            {/* Step 3: Run */}
            <div className="space-y-3">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-yapi-accent/20 border border-yapi-accent/40 flex items-center justify-center text-sm font-bold text-yapi-accent font-mono">
                  3
                </div>
                <h4 className="text-lg font-bold text-yapi-fg font-mono">Run it</h4>
              </div>
              <div className="ml-11 bg-yapi-bg-elevated border border-yapi-border/50 rounded-xl p-4">
                <pre className="text-sm text-yapi-fg font-mono overflow-x-auto">
{`# Run with explicit config file
yapi -c request.yaml

# Or use interactive file selector (git-tracked files)
yapi

# Search all YAML files
yapi --all`}
                </pre>
              </div>
            </div>

            {/* Advanced Example */}
            <div className="space-y-3 pt-4">
              <h4 className="text-lg font-bold text-yapi-fg text-center font-mono">
                More Examples
              </h4>
              <div className="grid md:grid-cols-2 gap-4">
                {/* HTTP POST */}
                <div className="bg-yapi-bg-elevated border border-yapi-border/50 rounded-xl p-4">
                  <div className="text-xs text-yapi-fg-muted mb-2 font-mono">HTTP POST</div>
                  <pre className="text-xs text-yapi-fg font-mono leading-relaxed overflow-x-auto">
{`url: https://api.example.com/users
method: POST
content_type: application/json
body:
  name: "Jane Doe"
  email: "jane@example.com"`}
                  </pre>
                </div>

                {/* gRPC */}
                <div className="bg-yapi-bg-elevated border border-yapi-border/50 rounded-xl p-4">
                  <div className="text-xs text-yapi-fg-muted mb-2 font-mono">gRPC</div>
                  <pre className="text-xs text-yapi-fg font-mono leading-relaxed overflow-x-auto">
{`url: grpc://localhost:50051
service: user.UserService
method: GetUser
body:
  user_id: 123`}
                  </pre>
                </div>
              </div>
            </div>
          </div>

          {/* CTA */}
          <div className="text-center space-y-6">
            <h3 className="text-2xl font-bold text-yapi-fg font-mono">
              Ready to try it?
            </h3>
            <p className="text-lg text-yapi-fg-muted">
              Test it in the{" "}
              <Link href="/playground" className="text-yapi-accent hover:underline font-semibold">
                web playground
              </Link>{" "}
              or clone the repo to get started locally
            </p>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="relative px-6 py-8 border-t border-yapi-border/50 bg-yapi-bg-elevated/50 backdrop-blur-sm">
        <div className="max-w-6xl mx-auto text-center text-sm text-yapi-fg-subtle font-mono">
          <p>Built with 🐑 · Open source · MIT License</p>
        </div>
      </footer>

      <style>{`
        @keyframes shimmer {
          0% { transform: translateX(-100%); }
          100% { transform: translateX(100%); }
        }
        .animate-shimmer {
          animation: shimmer 2s ease-in-out infinite;
        }
      `}</style>
    </div>
  );
}
