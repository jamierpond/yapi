import type {
  MadeaConfigWithSeo,
  ArticleViewProps,
  FileBrowserViewProps,
  FileInfo,
} from "madea-blog-core";
import { LocalFsDataProvider } from "madea-blog-core/providers/local-fs";
import { generateMetadataForIndex, generateMetadataForArticle } from "madea-blog-core";
import Link from "next/link";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeHighlight from "rehype-highlight";
import path from "path";
import fs from "fs";
import { execSync } from "child_process";

// Get version info at build time (uses Vercel env vars, falls back to git for local dev)
function getVersionInfo(): { semver: string; commit: string } {
  // Vercel provides VERCEL_GIT_COMMIT_SHA
  const commit = process.env.VERCEL_GIT_COMMIT_SHA?.slice(0, 7)
    ?? tryExec("git rev-parse --short HEAD")
    ?? "unknown";

  // For semver, try git tags (Vercel doesn't provide this)
  const tag = tryExec("git describe --tags --abbrev=0") ?? "0.0.0";
  const semver = tag.replace(/^v/, "");

  return { semver, commit };
}

function tryExec(cmd: string): string | null {
  try {
    const repoRoot = path.join(process.cwd(), "..");
    return execSync(cmd, { cwd: repoRoot, encoding: "utf-8", stdio: ["pipe", "pipe", "ignore"] }).trim();
  } catch {
    return null;
  }
}

const versionInfo = getVersionInfo();

function VersionFooter() {
  return (
    <p className="text-sm text-yapi-fg-muted">
      Docs generated from{" "}
      <a
        href={`https://github.com/jamierpond/yapi/tree/${versionInfo.commit}`}
        target="_blank"
        rel="noopener noreferrer"
        className="text-yapi-accent hover:underline"
      >
        yapi v{versionInfo.semver}-{versionInfo.commit}
      </a>
    </p>
  );
}

import Navbar from "@/app/components/Navbar";
import "highlight.js/styles/github-dark.css";

function DocsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen flex flex-col bg-yapi-bg relative overflow-hidden font-sans text-yapi-fg selection:bg-yapi-accent selection:text-white">
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute inset-0 bg-[linear-gradient(to_right,#80808012_1px,transparent_1px),linear-gradient(to_bottom,#80808012_1px,transparent_1px)] bg-[size:24px_24px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_100%)]"></div>
        <div className="absolute top-[-20%] left-[-10%] w-[50rem] h-[50rem] bg-yapi-accent/10 rounded-full blur-[120px] opacity-30"></div>
        <div className="absolute bottom-[-20%] right-[-10%] w-[40rem] h-[40rem] bg-indigo-500/10 rounded-full blur-[120px] opacity-20"></div>
      </div>
      <Navbar />
      <main className="flex-1 relative z-10 flex flex-col items-center pt-12 pb-32 px-6">
        {children}
      </main>
    </div>
  );
}

function ArticleView({ article }: ArticleViewProps) {
  return (
    <DocsLayout>
      <article className="max-w-3xl w-full">
        <Link
          href="/docs"
          className="inline-flex items-center gap-2 text-yapi-fg-muted hover:text-yapi-accent transition-colors mb-8"
        >
          <svg
            className="w-4 h-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M15 19l-7-7 7-7"
            />
          </svg>
          Back to Docs
        </Link>

        <header className="mb-12">
          <h1 className="text-4xl md:text-5xl font-bold tracking-tight mb-4">
            {article.title}
          </h1>
        </header>

        <div className="prose prose-invert prose-lg max-w-none prose-headings:text-yapi-fg prose-headings:font-bold prose-p:text-yapi-fg-muted prose-a:text-yapi-accent prose-a:no-underline hover:prose-a:underline prose-strong:text-yapi-fg prose-code:text-yapi-accent prose-code:bg-yapi-bg-elevated prose-code:px-1.5 prose-code:py-0.5 prose-code:rounded prose-code:before:content-none prose-code:after:content-none prose-pre:bg-[#1e1e1e] prose-pre:border prose-pre:border-yapi-border prose-blockquote:border-l-yapi-accent prose-blockquote:text-yapi-fg-muted prose-li:text-yapi-fg-muted prose-li:marker:text-yapi-accent">
          <Markdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeHighlight]}>{article.content}</Markdown>
        </div>

        <footer className="mt-16 pt-8 border-t border-yapi-border">
          <VersionFooter />
        </footer>
      </article>
    </DocsLayout>
  );
}

function extractDescription(content: string): string {
  // Get the first non-empty line after the title (## heading)
  const lines = content.split('\n');
  let foundTitle = false;
  for (const line of lines) {
    if (line.startsWith('## ')) {
      foundTitle = true;
      continue;
    }
    if (foundTitle && line.trim() && !line.startsWith('```')) {
      return line.trim();
    }
  }
  return '';
}

function FileBrowserView({ articles }: FileBrowserViewProps) {
  return (
    <DocsLayout>
      <div className="max-w-4xl w-full">
        <header className="text-center mb-16">
          <h1 className="text-5xl md:text-6xl font-bold tracking-tight mb-4">
            <span className="bg-gradient-to-r from-yapi-accent via-orange-300 to-yapi-accent bg-clip-text text-transparent">
              CLI Documentation
            </span>
          </h1>
          <p className="text-xl text-yapi-fg-muted max-w-xl mx-auto mb-4">
            Auto-generated documentation for yapi CLI commands
          </p>
          <VersionFooter />
        </header>

        {articles.length === 0 ? (
          <div className="text-center py-16">
            <p className="text-yapi-fg-muted">No docs yet. Run <code className="text-yapi-accent">go run scripts/gendocs.go</code> to generate.</p>
          </div>
        ) : (
          <div className="space-y-4">
            {articles.map((article: FileInfo) => {
              const description = extractDescription(article.content);
              return (
                <Link
                  key={article.sha}
                  href={`/docs/${article.path.replace(/\.md$/, "")}`}
                  className="block group p-5 rounded-xl border border-yapi-border bg-yapi-bg-elevated/20 hover:bg-yapi-bg-elevated/40 hover:border-yapi-accent/50 transition-all duration-300"
                >
                  <h2 className="text-lg font-bold group-hover:text-yapi-accent transition-colors font-mono">
                    {article.title}
                  </h2>
                  {description && (
                    <p className="text-sm text-yapi-fg-muted mt-1">
                      {description}
                    </p>
                  )}
                </Link>
              );
            })}
          </div>
        )}
      </div>
    </DocsLayout>
  );
}

function NoRepoFoundView() {
  return (
    <DocsLayout>
      <div className="text-center py-16">
        <h1 className="text-2xl font-bold text-yapi-fg mb-4">
          Docs Not Available
        </h1>
        <p className="text-yapi-fg-muted">
          Could not load documentation. Run <code className="text-yapi-accent">go run scripts/gendocs.go</code> to generate.
        </p>
        <Link
          href="/"
          className="inline-block mt-6 px-6 py-2 rounded-lg border border-yapi-border hover:border-yapi-accent transition-colors"
        >
          Go Home
        </Link>
      </div>
    </DocsLayout>
  );
}

function LandingView() {
  return (
    <DocsLayout>
      <h1 className="text-4xl font-bold">Welcome to the Docs</h1>
    </DocsLayout>
  );
}

const contentDir = path.join(process.cwd(), "app/_docs");

// Ensure directory exists for LocalFsDataProvider (which uses simple-git)
if (!fs.existsSync(contentDir)) {
  fs.mkdirSync(contentDir, { recursive: true });
}

export const docsDataProvider = new LocalFsDataProvider({
  contentDir,
  authorName: "yapi",
  sourceUrl: "https://github.com/jamierpond/yapi",
});

import { OG_BASE_URL } from "@/app/lib/constants";

const BASE_URL = "https://yapi.run";

const SEO_CONFIG = {
  baseUrl: BASE_URL,
  siteName: "yapi",
  defaultDescription: "CLI documentation for yapi - the API development toolkit",
  authorName: "yapi",
  authorUrl: BASE_URL,
} as const;

export function createDocsConfig(): MadeaConfigWithSeo {
  return {
    dataProvider: docsDataProvider,
    username: "yapi",
    fileBrowserView: FileBrowserView,
    articleView: ArticleView,
    noRepoFoundView: NoRepoFoundView,
    landingView: LandingView,
    seo: SEO_CONFIG,
    basePath: "/docs",
  };
}

// Re-export helpers bound to config for use in pages
export async function generateDocsMetadata() {
  const config = createDocsConfig();
  const metadata = await generateMetadataForIndex(config, {
    title: "CLI Documentation | yapi",
    description: "Auto-generated documentation for yapi CLI commands",
  });
  return {
    ...metadata,
    openGraph: {
      ...metadata.openGraph,
      images: [`${OG_BASE_URL}/og/docs?title=${encodeURIComponent("CLI Documentation")}`],
    },
  };
}

export async function generateDocsArticleMetadata(slug: string[]) {
  const config = createDocsConfig();
  const metadata = await generateMetadataForArticle(config, slug);
  const title =
    (metadata?.title && typeof metadata.title === "string"
      ? metadata.title
      : null) ?? slug.join("/");

  return {
    ...metadata,
    title,
    openGraph: {
      ...metadata?.openGraph,
      title,
      images: [`${OG_BASE_URL}/og/docs?title=${encodeURIComponent(title)}`],
    },
  };
}
