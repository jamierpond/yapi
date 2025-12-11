import Link from "next/link";
import CopyInstallButton from "./CopyInstallButton";
import LandingStyles from "./LandingStyles";
import Navbar from "./Navbar";
import { getTotalDownloads } from "@/app/lib/github";

async function getStats() {
  try {
    const [totalDownloads, releasesRes] = await Promise.all([
      getTotalDownloads(),
      fetch("https://api.github.com/repos/jamierpond/yapi/releases/latest", {
        next: { revalidate: 3600 },
      }),
    ]);

    const release = releasesRes.ok ? await releasesRes.json() : { tag_name: null };

    return {
      totalDownloads: totalDownloads || 0,
      latestVersion: release.tag_name || null,
    };
  } catch {
    return { totalDownloads: 0, latestVersion: null };
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
          {stats.totalDownloads > 0 && (
            <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full border border-yapi-border bg-yapi-bg-elevated/50 backdrop-blur-sm shadow-sm">
              <span className="text-xs font-mono text-yapi-fg-muted">
                {stats.totalDownloads.toLocaleString()} downloads
              </span>
            </div>
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
            Define API requests in YAML. Run them from your terminal. HTTP, gRPC, GraphQL. Commit to git. No Postman. No Insomnia.
          </p>

          <div className="flex flex-col justify-center items-center gap-4 pt-8 animate-fade-in-up delay-75 w-full max-w-xl mx-auto">
            <CopyInstallButton />
            <Link
              href="/playground"
              className="px-8 py-3 rounded-xl border border-yapi-border bg-yapi-bg-elevated/40 text-yapi-fg font-bold hover:bg-yapi-bg-elevated hover:border-yapi-accent/50 transition-all active:scale-[0.98] w-full sm:w-auto text-center"
            >
              Try Online
            </Link>
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
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-yapi-fg-muted font-mono">create-user.yapi.yml</span>
                  </div>
                  <div className="flex gap-1.5">
                    <div className="w-2.5 h-2.5 rounded-full bg-[#ff5f56]"></div>
                    <div className="w-2.5 h-2.5 rounded-full bg-[#ffbd2e]"></div>
                    <div className="w-2.5 h-2.5 rounded-full bg-[#27c93f]"></div>
                  </div>
                </div>
                <div className="p-6 font-mono text-sm leading-relaxed overflow-x-auto text-yapi-fg/90">
                  <div className="text-yapi-fg-subtle/50 mb-2"># Define your request in YAML</div>
                  <div><span className="text-yapi-accent">url</span>: <span className="text-orange-300">{"${BASE_URL}"}</span>/api/v1/users</div>
                  <div><span className="text-yapi-accent">method</span>: POST</div>
                  <div><span className="text-yapi-accent">headers</span>:</div>
                  <div>{"  "}<span className="text-blue-300">Authorization</span>: Bearer <span className="text-orange-300">{"${API_KEY}"}</span></div>
                  <div><span className="text-yapi-accent">body</span>:</div>
                  <div>{"  "}<span className="text-blue-300">username</span>: "dev_sheep"</div>
                  <div>{"  "}<span className="text-blue-300">role</span>: "admin"</div>
                  <div>{"  "}<span className="text-blue-300">features</span>:</div>
                  <div>{"    "}- "beta_access"</div>
                  <div>{"    "}- "unlimited_wool"</div>
                </div>
              </div>

              {/* Right Pane: The Execution (Terminal) */}
              <div className="flex-1 bg-[#0c0c0c] flex flex-col">
                 <div className="bg-[#1a1a1a] px-4 py-2 flex items-center border-b border-black/20">
                    <span className="text-xs text-yapi-fg-subtle font-mono">zsh — yapi run</span>
                 </div>
                 <div className="p-6 font-mono text-sm leading-relaxed overflow-x-auto relative h-full">
                    {/* Scanline */}
                    <div className="absolute inset-0 bg-[linear-gradient(rgba(18,16,16,0)_50%,rgba(0,0,0,0.1)_50%)] bg-[size:100%_4px] pointer-events-none opacity-20"></div>

                    <div className="text-yapi-fg-muted mb-4">
                      $ yapi run create-user.yapi.yml
                    </div>

                    <div>
                      <span className="text-yapi-success font-bold">201 Created</span> <span className="text-yapi-fg-subtle text-xs ml-2">124ms</span>
                      <br/><br/>
                      <span className="text-yellow-500">{`{`}</span><br/>
                      {"  "}<span className="text-blue-400">"id"</span>: <span className="text-green-400">"usr_8a92b"</span>,<br/>
                      {"  "}<span className="text-blue-400">"status"</span>: <span className="text-green-400">"active"</span>,<br/>
                      {"  "}<span className="text-blue-400">"message"</span>: <span className="text-green-400">"Welcome to the flock."</span><br/>
                      <span className="text-yellow-500">{`}`}</span>
                      <span className="inline-block w-2 h-4 bg-yapi-accent ml-2 align-middle animate-pulse"></span>
                    </div>
                 </div>
              </div>
           </div>
        </div>

        {/* Feature Grid */}
        <div className="max-w-6xl w-full mx-auto grid md:grid-cols-3 gap-8 mt-32">
          <FeatureCard
            icon="⚡"
            title="Go Native Speed"
            desc="Written in Go. Starts instantly. Uses minimal RAM. No Electron bloat, no loading spinners, no updates that move your buttons."
          />
          <FeatureCard
            icon="🤝"
            title="Team Friendly"
            desc="Review API changes in Pull Requests. Diff your request bodies. Merge conflicts are just text conflicts. True collaboration."
          />
          <FeatureCard
            icon="🧠"
            title="Built-in LSP"
            desc="Full Language Server with autocompletion, real-time validation, and hover info. Works with Neovim, VS Code, and any LSP-compatible editor."
          />
        </div>

        {/* Advanced Feature Highlight */}
        <div className="max-w-4xl w-full mx-auto mt-32 mb-16 border-t border-yapi-border/50 pt-16">
           <div className="flex flex-col md:flex-row gap-12 items-center">
              <div className="flex-1 space-y-4">
                <div className="inline-block px-3 py-1 rounded bg-yapi-success/10 text-yapi-success text-xs font-bold uppercase tracking-wider">
                  New
                </div>
                <h3 className="text-3xl font-bold">Request Chaining & Assertions</h3>
                <p className="text-yapi-fg-muted leading-relaxed">
                  Build complex workflows without a GUI scripting engine.
                  Chain requests declaratively, pass data between steps, and validate responses
                  with JQ-powered assertions. Perfect for auth flows and integration tests.
                </p>
              </div>
              <div className="flex-1 w-full">
                 <div className="bg-[#1e1e1e] border border-yapi-border rounded-lg p-4 font-mono text-xs shadow-lg overflow-x-auto">
                   <pre className="text-yapi-fg-muted whitespace-pre">
                     <code>
{`chain:
  `}<span className="text-yapi-fg-subtle"># Step 1: Login</span>{`
  - `}<span className="text-yapi-accent">name</span>{`: auth
    `}<span className="text-yapi-accent">url</span>{`: /login
    `}<span className="text-yapi-accent">body</span>{`: { user: "me" }
    `}<span className="text-yapi-accent">expect</span>{`:
      `}<span className="text-blue-300">status</span>{`: 200
      `}<span className="text-blue-300">assert</span>{`:
        - `}<span className="text-green-400">.token != null</span>{`

  `}<span className="text-yapi-fg-subtle"># Step 2: Use Token</span>{`
  - `}<span className="text-yapi-accent">name</span>{`: profile
    `}<span className="text-yapi-accent">url</span>{`: /me
    `}<span className="text-yapi-accent">headers</span>{`:
      `}<span className="text-blue-300">Authorization</span>{`: `}<span className="text-orange-300">{'${auth.token}'}</span>{`
    `}<span className="text-yapi-accent">expect</span>{`:
      `}<span className="text-blue-300">assert</span>{`:
        - `}<span className="text-green-400">.email != null</span>
                     </code>
                   </pre>
                 </div>
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

      <LandingStyles />
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
