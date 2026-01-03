import Link from "next/link";
import CopyInstallButton from "./CopyInstallButton";
import Navbar from "./Navbar";
import { getTotalDownloads, getVSCodeInstalls, getOpenVSXDownloads, getGitHubStats } from "@/app/lib/github";

async function getStats() {
  try {
    const FIVE_MINUTES_SECONDS = 300;
    const [totalDownloads, releasesRes, vscodeInstalls, openVSXDownloads, githubStats] = await Promise.all([
      getTotalDownloads(),
      fetch("https://api.github.com/repos/jamierpond/yapi/releases/latest", {
        next: { revalidate: FIVE_MINUTES_SECONDS },
      }),
      getVSCodeInstalls(),
      getOpenVSXDownloads(),
      getGitHubStats(),
    ]);

    const release = releasesRes.ok ? await releasesRes.json() : { tag_name: null };

    const cliDownloads = totalDownloads || 0;
    const extensionInstalls = (vscodeInstalls || 0) + (openVSXDownloads || 0);

    return {
      latestVersion: release.tag_name || null,
      githubStars: githubStats.stars || 0,
      githubForks: githubStats.forks || 0,
      cliDownloads,
      extensionInstalls,
      totalInstalls: cliDownloads + extensionInstalls,
    };
  } catch {
    return { latestVersion: null, githubStars: 0, githubForks: 0, cliDownloads: 0, extensionInstalls: 0, totalInstalls: 0 };
  }
}

export default async function Landing() {
  const stats = await getStats();
  return (
    <div className="min-h-screen flex flex-col bg-yapi-bg relative overflow-hidden font-sans text-yapi-fg selection:bg-yapi-accent selection:text-white">
      {/* --- Fun Layer: Background Grid & Noise --- */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        {/* Moving Grid */}
        <div className="absolute inset-0 bg-[linear-gradient(to_right,#80808012_1px,transparent_1px),linear-gradient(to_bottom,#80808012_1px,transparent_1px)] bg-[size:24px_24px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_100%)]"></div>

        {/* Glowing Orbs */}
        <div className="absolute top-[-20%] left-[-10%] w-[50rem] h-[50rem] bg-yapi-accent/10 rounded-full blur-[120px] opacity-30 animate-pulse-slow"></div>
        <div className="absolute bottom-[-20%] right-[-10%] w-[40rem] h-[40rem] bg-indigo-500/10 rounded-full blur-[120px] opacity-20 animate-pulse-slow" style={{ animationDelay: '2s' }}></div>

        {/* Grain Overlay */}
        <div className="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20 mix-blend-soft-light"></div>
      </div>

      <Navbar />

      {/* Hero Section */}
      <main className="flex-1 relative z-10 flex flex-col items-center pt-20 pb-32 px-6">

        {/* Stats Bar */}
        <div className="mb-8 animate-fade-in-up flex flex-wrap justify-center gap-4">
          {stats.latestVersion && (
            <a
              href="https://github.com/jamierpond/yapi/releases/latest"
              className="inline-flex items-center gap-2 px-4 py-2 rounded-full border border-yapi-border bg-yapi-bg-elevated/50 backdrop-blur-sm shadow-sm hover:border-yapi-accent/50 transition-colors"
            >
              <span className="text-xs font-mono text-yapi-accent">{stats.latestVersion}</span>
            </a>
          )}
          <div className="inline-flex items-center gap-3 px-4 py-2 rounded-full border border-yapi-border bg-yapi-bg-elevated/50 backdrop-blur-sm shadow-sm">
            <div className="flex h-2 w-2 relative">
              <span className="relative inline-flex rounded-full h-2 w-2 bg-yapi-success"></span>
            </div>
            <span className="text-xs font-mono text-yapi-fg-muted">
              No Cloud Sync. No Forced Login.
            </span>
          </div>
        </div>

        <div className="max-w-5xl w-full text-center space-y-6 mb-16">
          <h1 className="text-5xl md:text-7xl font-bold tracking-tight leading-[1.1]">
            YAML in.<br className="hidden md:block" />
            <span className="bg-gradient-to-r from-yapi-accent via-orange-300 to-yapi-accent bg-clip-text text-transparent animate-shine bg-[length:200%_auto]">
              HTTP out.
            </span>
          </h1>

          <p className="text-xl text-yapi-fg-muted max-w-xl mx-auto leading-relaxed">
            Chain API requests in YAML. Switch environments with a flag. Run complex workflows from your terminal. Commit to git. No Electron apps.
          </p>

          <div className="flex flex-col justify-center items-center gap-4 pt-8 animate-fade-in-up delay-75 w-full max-w-xl mx-auto">
            <CopyInstallButton />
            <div className="flex flex-col sm:flex-row gap-4 w-full sm:w-auto">
              <Link
                href="/playground"
                className="px-8 py-3 rounded-xl border border-yapi-border bg-yapi-bg-elevated/40 text-yapi-fg font-bold hover:bg-yapi-bg-elevated hover:border-yapi-accent/50 transition-all active:scale-[0.98] w-full sm:w-auto text-center"
              >
                Try Online
              </Link>
              <a
                href="https://github.com/jamierpond/yapi"
                target="_blank"
                rel="noopener noreferrer"
                className="px-8 py-3 rounded-xl border border-yapi-border bg-yapi-bg-elevated/40 text-yapi-fg font-bold hover:bg-yapi-bg-elevated hover:border-yapi-accent/50 transition-all active:scale-[0.98] w-full sm:w-auto text-center flex items-center justify-center gap-2"
              >
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                  <path fillRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clipRule="evenodd" />
                </svg>
                Star on GitHub
              </a>
            </div>
          </div>

          <p className="mt-6 text-xs text-yapi-fg-subtle opacity-50 font-mono text-center">
            Requires curl (macOS/Linux) or PowerShell (Windows)
          </p>
        </div>

        {/* Hero Visual: The Split Pane Terminal */}
        <div className="max-w-6xl w-full relative group perspective-1000 animate-fade-in-up delay-100">
           {/* Glow behind terminal */}
           <div className="absolute -inset-1 bg-gradient-to-b from-yapi-accent/20 to-transparent rounded-2xl blur-xl opacity-20 group-hover:opacity-30 transition duration-1000"></div>

           <div className="relative bg-[#1e1e1e] border border-yapi-border rounded-xl shadow-2xl overflow-hidden flex flex-col md:flex-row min-h-[400px]">

              {/* Left Pane: The Config (Editor) */}
              <div className="flex-1 border-r border-white/5 flex flex-col">
                <div className="bg-[#252526] px-4 py-2 flex items-center justify-between border-b border-black/20">
                  <div className="flex gap-1.5">
                    <div className="w-2.5 h-2.5 rounded-full bg-[#ff5f56]"></div>
                    <div className="w-2.5 h-2.5 rounded-full bg-[#ffbd2e]"></div>
                    <div className="w-2.5 h-2.5 rounded-full bg-[#27c93f]"></div>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-yapi-fg-muted font-mono">auth-flow.yapi.yml</span>
                  </div>
                </div>
                <div className="p-6 font-mono text-sm leading-relaxed overflow-x-auto text-yapi-fg/90 whitespace-pre">
                  <div><span className="text-yapi-accent">yapi</span>: v1</div>
                  <div className="text-yapi-fg-subtle/50 mb-2"># Chain requests together</div>
                  <div><span className="text-yapi-accent">chain</span>:</div>
                  <div>  - <span className="text-yapi-accent">name</span>: login</div>
                  <div>    <span className="text-yapi-accent">url</span>: <span className="text-orange-300">{"${url}"}</span>/auth/login</div>
                  <div>    <span className="text-yapi-accent">method</span>: POST</div>
                  <div>    <span className="text-yapi-accent">body</span>:</div>
                  <div>      <span className="text-blue-300">user</span>: "dev_sheep"</div>
                  <div>    <span className="text-yapi-accent">expect</span>:</div>
                  <div>      <span className="text-blue-300">assert</span>:</div>
                  <div>        - <span className="text-green-400">.token != null</span></div>
                  <div className="mt-2">  - <span className="text-yapi-accent">name</span>: get_profile</div>
                  <div>    <span className="text-yapi-accent">url</span>: <span className="text-orange-300">{"${url}"}</span>/me</div>
                  <div>    <span className="text-yapi-accent">headers</span>:</div>
                  <div>      <span className="text-blue-300">Auth</span>: Bearer <span className="text-orange-300">{"${login.token}"}</span></div>
                </div>
              </div>

              {/* Right Pane: The Execution (Terminal) */}
              <div className="flex-1 bg-[#0c0c0c] flex flex-col">
                 <div className="bg-[#1a1a1a] px-4 py-2 flex items-center border-b border-black/20">
                    <span className="text-xs text-yapi-fg-subtle font-mono">zsh — yapi run</span>
                 </div>
                 <div className="p-6 font-mono text-sm leading-relaxed overflow-x-auto relative h-full whitespace-pre">
                    {/* Scanline */}
                    <div className="absolute inset-0 bg-[linear-gradient(rgba(18,16,16,0)_50%,rgba(0,0,0,0.1)_50%)] bg-[size:100%_4px] pointer-events-none opacity-20"></div>

                    <div className="text-yapi-fg-muted mb-3">
                      $ yapi run auth-flow.yapi.yml -e prod
                    </div>

                    <div className="text-yapi-fg-muted text-xs mb-2">
                      Running step 1: login...<br/>
                      Running step 2: get_profile...
                    </div>

                    <div className="mb-3">
                      <div className="text-yapi-fg-subtle mb-1">--- Step 2: get_profile ---</div>
                      <span className="text-yellow-500">{`{`}</span><br/>
                      {"  "}<span className="text-blue-400">"name"</span>: <span className="text-green-400">"Dev Sheep"</span>,<br/>
                      {"  "}<span className="text-blue-400">"email"</span>: <span className="text-green-400">"dev@example.com"</span>,<br/>
                      {"  "}<span className="text-blue-400">"role"</span>: <span className="text-green-400">"admin"</span><br/>
                      <span className="text-yellow-500">{`}`}</span>
                    </div>

                    <div className="text-yapi-fg-subtle text-xs mb-2">
                      URL: https://api.example.com/me<br/>
                      Time: 67ms<br/>
                      Size: 78B
                    </div>

                    <div className="text-yapi-success text-sm">
                      Chain completed successfully.
                      <span className="inline-block w-2 h-4 bg-yapi-accent ml-2 align-middle animate-pulse"></span>
                    </div>
                 </div>
              </div>
           </div>
        </div>

        {/* Vanity Metrics */}
        <div className="w-full px-6 mt-20">
          <div className="grid grid-cols-2 sm:grid-cols-5 gap-3 max-w-3xl mx-auto">
            <MetricCard
              href="https://github.com/jamierpond/yapi/stargazers"
              value={stats.githubStars}
              label="Stars"
            />
            <MetricCard
              href="https://github.com/jamierpond/yapi/forks"
              value={stats.githubForks}
              label="Forks"
            />
            <MetricCard
              href="https://github.com/jamierpond/yapi/releases"
              value={stats.cliDownloads}
              label="CLI"
            />
            <MetricCard
              href="https://marketplace.visualstudio.com/items?itemName=yapi.yapi-extension"
              value={stats.extensionInstalls}
              label="Extension"
              hoverColor="blue"
            />
            <div className="p-4 rounded-xl border border-yapi-border bg-gradient-to-br from-yapi-accent/10 to-purple-500/10 text-center col-span-2 sm:col-span-1">
              <div className="text-2xl font-bold bg-gradient-to-r from-yapi-accent to-purple-400 bg-clip-text text-transparent">
                {stats.totalInstalls.toLocaleString()}
              </div>
              <div className="text-[10px] text-yapi-fg-muted mt-1 font-mono uppercase tracking-wide">Total</div>
            </div>
          </div>
        </div>

        {/* Feature Grid */}
        <div className="max-w-6xl w-full mx-auto grid md:grid-cols-3 gap-8 mt-32">
          <FeatureCard
            icon="🔗"
            title="Request Chaining"
            desc="Chain requests together declaratively. Pass data between steps. Validate with assertions. Build auth flows and integration tests without a separate framework."
          />
          <FeatureCard
            icon="🌍"
            title="Environment Configs"
            desc="One yapi.config.yml to manage dev, staging, and prod. Switch with a flag. No duplicate files. Load secrets from shell env. Perfect for teams and CI/CD."
          />
          <FeatureCard
            icon="⚡"
            title="Go Native Speed"
            desc="Written in Go. Starts instantly. Uses minimal RAM. No Electron bloat, no loading spinners, no updates that move your buttons. Just a fast, reliable CLI tool."
          />
        </div>

        {/* VS Code / Cursor Showcase */}
        <div className="max-w-6xl w-full mx-auto mt-32">
          <div className="text-center mb-12">
            <h2 className="text-3xl md:text-4xl font-bold mb-4">
              First-class{" "}
              <span className="bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
                VS Code & Cursor
              </span>{" "}
              support
            </h2>
            <p className="text-yapi-fg-muted max-w-2xl mx-auto">
              Run requests with Cmd+Enter. See responses inline. Full LSP with autocompletion, validation, and hover info. No context switching.
            </p>
          </div>

          <div className="relative group">
            {/* Glow effect */}
            <div className="absolute -inset-1 bg-gradient-to-r from-blue-500/20 via-purple-500/20 to-blue-500/20 rounded-2xl blur-xl opacity-30 group-hover:opacity-50 transition duration-500"></div>

            <div className="relative rounded-xl overflow-hidden border border-yapi-border shadow-2xl">
              <img
                src="/image.png"
                alt="yapi VS Code extension showing inline request execution with response panel"
                className="w-full h-auto"
              />
            </div>
          </div>

          <div className="flex flex-wrap justify-center gap-4 mt-8">
            <a
              href="https://marketplace.visualstudio.com/items?itemName=yapi.yapi-extension"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 px-6 py-3 rounded-xl border border-yapi-border bg-yapi-bg-elevated/40 text-yapi-fg font-medium hover:bg-yapi-bg-elevated hover:border-blue-500/50 transition-all"
            >
              <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
                <path d="M23.15 2.587L18.21.21a1.494 1.494 0 0 0-1.705.29l-9.46 8.63-4.12-3.128a.999.999 0 0 0-1.276.057L.327 7.261A1 1 0 0 0 .326 8.74L3.899 12 .326 15.26a1 1 0 0 0 .001 1.479L1.65 17.94a.999.999 0 0 0 1.276.057l4.12-3.128 9.46 8.63a1.492 1.492 0 0 0 1.704.29l4.942-2.377A1.5 1.5 0 0 0 24 20.06V3.939a1.5 1.5 0 0 0-.85-1.352zm-5.146 14.861L10.826 12l7.178-5.448v10.896z"/>
              </svg>
              VS Code Marketplace
            </a>
            <a
              href="https://open-vsx.org/extension/yapi/yapi-extension"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 px-6 py-3 rounded-xl border border-yapi-border bg-yapi-bg-elevated/40 text-yapi-fg font-medium hover:bg-yapi-bg-elevated hover:border-purple-500/50 transition-all"
            >
              Open VSX (Cursor)
            </a>
          </div>
        </div>

        {/* Additional Features */}
        <div className="max-w-6xl w-full mx-auto mt-32 mb-16">
           <div className="grid md:grid-cols-2 gap-8">
              <div className="p-8 rounded-2xl border border-yapi-border bg-yapi-bg-elevated/20">
                <div className="text-3xl mb-4">🧠</div>
                <h3 className="text-xl font-bold mb-3">Built-in LSP</h3>
                <p className="text-yapi-fg-muted leading-relaxed text-sm">
                  Full Language Server with autocompletion, real-time validation, and hover info. Works with Neovim, VS Code, Cursor, and any LSP-compatible editor.
                </p>
              </div>
              <div className="p-8 rounded-2xl border border-yapi-border bg-yapi-bg-elevated/20">
                <div className="text-3xl mb-4">🤝</div>
                <h3 className="text-xl font-bold mb-3">Team Friendly</h3>
                <p className="text-yapi-fg-muted leading-relaxed text-sm">
                  Review API changes in Pull Requests. Diff your request bodies. Merge conflicts are just text conflicts. True collaboration without shared cloud workspaces.
                </p>
              </div>
           </div>
        </div>

      </main>

      {/* Footer */}
      <footer className="border-t border-yapi-border/50 bg-yapi-bg-elevated/30 py-12 px-6">
        <div className="max-w-7xl mx-auto flex flex-col md:flex-row justify-between items-center gap-6">
          <div className="text-yapi-fg-muted text-sm font-mono opacity-60">
             Built for developers who prefer the terminal.
          </div>
          <div className="flex gap-6">
            <a href="https://github.com/jamierpond/yapi" className="text-yapi-fg-subtle hover:text-yapi-accent transition-colors text-sm">Source Code</a>
            <a href="/docs" className="text-yapi-fg-subtle hover:text-yapi-accent transition-colors text-sm">Documentation</a>
          </div>
        </div>
      </footer>
    </div>
  );
}

function FeatureCard({ icon, title, desc }: { icon: string, title: string, desc: string }) {
  return (
    <div className="group p-8 rounded-2xl bg-yapi-bg-elevated/20 border border-yapi-border hover:bg-yapi-bg-elevated/40 transition-all duration-300">
      <div className="h-12 w-12 rounded-lg bg-yapi-bg-subtle flex items-center justify-center mb-6 text-2xl shadow-inner group-hover:scale-110 group-hover:-rotate-3 transition-transform duration-300">
        {icon}
      </div>
      <h3 className="text-xl font-bold mb-3 group-hover:text-yapi-accent transition-colors">{title}</h3>
      <p className="text-yapi-fg-muted leading-relaxed text-sm">
        {desc}
      </p>
    </div>
  );
}

function MetricCard({
  href,
  value,
  label,
  hoverColor = "accent"
}: {
  href: string;
  value: number;
  label: string;
  hoverColor?: "accent" | "blue" | "purple";
}) {
  if (value <= 0) return null;

  const baseClasses = "p-4 rounded-xl border border-yapi-border bg-yapi-bg-elevated/20 hover:bg-yapi-bg-elevated/40 transition-all text-center";
  const colorClasses = {
    accent: "hover:border-yapi-accent/50",
    blue: "hover:border-blue-500/50",
    purple: "hover:border-purple-500/50",
  };
  const textColorClasses = {
    accent: "group-hover:text-yapi-accent",
    blue: "group-hover:text-blue-400",
    purple: "group-hover:text-purple-400",
  };

  return (
    <a href={href} className={`group ${baseClasses} ${colorClasses[hoverColor]}`}>
      <div className={`text-2xl font-bold text-yapi-fg transition-colors ${textColorClasses[hoverColor]}`}>
        {value.toLocaleString()}
      </div>
      <div className="text-[10px] text-yapi-fg-muted mt-1 font-mono uppercase tracking-wide">{label}</div>
    </a>
  );
}
