import type {
  Config,
  ArticleViewProps,
  FileBrowserViewProps,
  NoRepoFoundViewProps,
  FileInfo,
} from "madea-blog-core";
import { LocalFsDataProvider } from "madea-blog-core/providers/local-fs";
import Link from "next/link";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import path from "path";
import Navbar from "@/app/components/Navbar";

function BlogLayout({ children }: { children: React.ReactNode }) {
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
    <BlogLayout>
      <article className="max-w-3xl w-full">
        <Link
          href="/blog"
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
          Back to Blog
        </Link>

        <header className="mb-12">
          <h1 className="text-4xl md:text-5xl font-bold tracking-tight mb-4">
            {article.title}
          </h1>
          <div className="flex items-center gap-4 text-sm text-yapi-fg-muted">
            <time dateTime={article.commitInfo.date}>
              {new Date(article.commitInfo.date).toLocaleDateString("en-US", {
                year: "numeric",
                month: "long",
                day: "numeric",
              })}
            </time>
            {article.commitInfo.authorName && (
              <>
                <span className="text-yapi-border">|</span>
                <span>{article.commitInfo.authorName}</span>
              </>
            )}
          </div>
        </header>

        <div className="prose prose-invert prose-lg max-w-none prose-headings:text-yapi-fg prose-headings:font-bold prose-p:text-yapi-fg-muted prose-a:text-yapi-accent prose-a:no-underline hover:prose-a:underline prose-strong:text-yapi-fg prose-code:text-yapi-accent prose-code:bg-yapi-bg-elevated prose-code:px-1.5 prose-code:py-0.5 prose-code:rounded prose-code:before:content-none prose-code:after:content-none prose-pre:bg-[#1e1e1e] prose-pre:border prose-pre:border-yapi-border prose-blockquote:border-l-yapi-accent prose-blockquote:text-yapi-fg-muted prose-li:text-yapi-fg-muted prose-li:marker:text-yapi-accent">
          <Markdown remarkPlugins={[remarkGfm]}>{article.content}</Markdown>
        </div>
      </article>
    </BlogLayout>
  );
}

function FileBrowserView({ articles }: FileBrowserViewProps) {
  return (
    <BlogLayout>
      <div className="max-w-4xl w-full">
        <header className="text-center mb-16">
          <h1 className="text-5xl md:text-6xl font-bold tracking-tight mb-4">
            <span className="bg-gradient-to-r from-yapi-accent via-orange-300 to-yapi-accent bg-clip-text text-transparent">
              Blog
            </span>
          </h1>
          <p className="text-xl text-yapi-fg-muted max-w-xl mx-auto">
            Updates, tutorials, and thoughts about yapi
          </p>
        </header>

        {articles.length === 0 ? (
          <div className="text-center py-16">
            <p className="text-yapi-fg-muted">No posts yet. Stay tuned.</p>
          </div>
        ) : (
          <div className="space-y-6">
            {articles.map((article: FileInfo) => (
              <Link
                key={article.sha}
                href={`/blog/${article.path.replace(/\.md$/, "")}`}
                className="block group p-6 rounded-xl border border-yapi-border bg-yapi-bg-elevated/20 hover:bg-yapi-bg-elevated/40 hover:border-yapi-accent/50 transition-all duration-300"
              >
                <h2 className="text-xl font-bold mb-2 group-hover:text-yapi-accent transition-colors">
                  {article.title}
                </h2>
                <time
                  dateTime={article.commitInfo.date}
                  className="text-sm text-yapi-fg-subtle"
                >
                  {new Date(article.commitInfo.date).toLocaleDateString(
                    "en-US",
                    {
                      year: "numeric",
                      month: "long",
                      day: "numeric",
                    }
                  )}
                </time>
              </Link>
            ))}
          </div>
        )}
      </div>
    </BlogLayout>
  );
}

function NoRepoFoundView() {
  return (
    <BlogLayout>
      <div className="text-center py-16">
        <h1 className="text-2xl font-bold text-yapi-fg mb-4">
          Blog Not Available
        </h1>
        <p className="text-yapi-fg-muted">
          Could not load blog content. Please try again later.
        </p>
        <Link
          href="/"
          className="inline-block mt-6 px-6 py-2 rounded-lg border border-yapi-border hover:border-yapi-accent transition-colors"
        >
          Go Home
        </Link>
      </div>
    </BlogLayout>
  );
}

function LandingView() {
  return (
    <BlogLayout>
      <h1 className="text-4xl font-bold">Welcome to the Blog</h1>
    </BlogLayout>
  );
}

export function createBlogConfig(): Config {
  const contentDir = path.join(process.cwd(), "app/blog/_content");

  return {
    dataProvider: new LocalFsDataProvider({
      contentDir,
      authorName: "yapi",
      sourceUrl: "https://github.com/jamierpond/yapi",
    }),
    username: "yapi",
    fileBrowserView: FileBrowserView,
    articleView: ArticleView,
    noRepoFoundView: NoRepoFoundView,
    landingView: LandingView,
  };
}
