import type {
  MadeaConfigWithSeo,
  ArticleViewProps,
  FileBrowserViewProps,
  FileInfo,
} from "madea-blog-core";
import { generateArticleJsonLd, generateMetadataForIndex, generateMetadataForArticle } from "madea-blog-core";
import { GitHubDataProvider } from "madea-blog-core/providers/github"
import Link from "next/link";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeHighlight from "rehype-highlight";
import path from "path";
import Navbar from "@/app/components/Navbar";
import "highlight.js/styles/github-dark.css";

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

import { OG_BASE_URL } from "@/app/lib/constants";

const BASE_URL = "https://yapi.run";

function ArticleView({ article }: ArticleViewProps) {
  const jsonLd = generateArticleJsonLd(article, {
    baseUrl: BASE_URL,
    blogPath: "/blog",
    authorName: "yapi",
    authorUrl: BASE_URL,
  });

  return (
    <BlogLayout>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
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
            {article.commitInfo.authorAvatarUrl && (
              <a
                href={`https://github.com/${article.commitInfo.authorUsername}`}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 hover:text-yapi-accent transition-colors"
              >
                <img
                  src={article.commitInfo.authorAvatarUrl}
                  alt={article.commitInfo.authorName || "Author"}
                  className="w-8 h-8 rounded-full"
                />
                <span>{article.commitInfo.authorName}</span>
              </a>
            )}
            <span className="text-yapi-border">|</span>
            <time dateTime={article.commitInfo.date}>
              {new Date(article.commitInfo.date).toLocaleDateString("en-US", {
                year: "numeric",
                month: "long",
                day: "numeric",
              })}
            </time>
            <span className="text-yapi-border">|</span>
            <a
              href={article.url}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1.5 hover:text-yapi-accent transition-colors"
            >
              <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                <path fillRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clipRule="evenodd" />
              </svg>
              View Source
            </a>
          </div>
        </header>

        <div className="prose prose-invert prose-lg max-w-none prose-headings:text-yapi-fg prose-headings:font-bold prose-p:text-yapi-fg-muted prose-a:text-yapi-accent prose-a:no-underline hover:prose-a:underline prose-strong:text-yapi-fg prose-code:text-yapi-accent prose-code:bg-yapi-bg-elevated prose-code:px-1.5 prose-code:py-0.5 prose-code:rounded prose-code:before:content-none prose-code:after:content-none prose-pre:bg-[#1e1e1e] prose-pre:border prose-pre:border-yapi-border prose-blockquote:border-l-yapi-accent prose-blockquote:text-yapi-fg-muted prose-li:text-yapi-fg-muted prose-li:marker:text-yapi-accent">
          <Markdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeHighlight]}>{article.content}</Markdown>
        </div>
      </article>
    </BlogLayout>
  );
}

function extractDescription(content: string): string {
  const lines = content.split('\n');
  let foundTitle = false;
  for (const line of lines) {
    if (line.startsWith('# ') || line.startsWith('## ')) {
      foundTitle = true;
      continue;
    }
    if (foundTitle && line.trim() && !line.startsWith('```') && !line.startsWith('#')) {
      const desc = line.trim();
      return desc.length > 150 ? desc.slice(0, 147) + '...' : desc;
    }
  }
  return '';
}

function FileBrowserView({ articles, sourceInfo }: FileBrowserViewProps) {
  return (
    <BlogLayout>
      <div className="max-w-4xl w-full">
        <header className="text-center mb-16">
          <h1 className="text-5xl md:text-6xl font-bold tracking-tight mb-4">
            <span className="bg-gradient-to-r from-yapi-accent via-orange-300 to-yapi-accent bg-clip-text text-transparent">
              Blog
            </span>
          </h1>
          <p className="text-xl text-yapi-fg-muted max-w-xl mx-auto mb-4">
            Updates, tutorials, and thoughts about yapi
          </p>
          <div className="flex items-center justify-center gap-4 text-sm text-yapi-fg-subtle">
            <a
              href={sourceInfo.sourceUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 hover:text-yapi-accent transition-colors"
            >
              <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                <path fillRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clipRule="evenodd" />
              </svg>
              View Source
            </a>
          </div>
        </header>

        {articles.length === 0 ? (
          <div className="text-center py-16">
            <p className="text-yapi-fg-muted">No posts yet. Stay tuned.</p>
          </div>
        ) : (
          <div className="space-y-4">
            {articles.map((article: FileInfo) => {
              const description = extractDescription(article.content);
              return (
                <Link
                  key={article.sha}
                  href={`/blog/${article.path.replace(/\.md$/, "")}`}
                  className="block group p-5 rounded-xl border border-yapi-border bg-yapi-bg-elevated/20 hover:bg-yapi-bg-elevated/40 hover:border-yapi-accent/50 transition-all duration-300"
                >
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      <h2 className="text-lg font-bold group-hover:text-yapi-accent transition-colors">
                        {article.title}
                      </h2>
                      {description && (
                        <p className="text-sm text-yapi-fg-muted mt-1 line-clamp-2">
                          {description}
                        </p>
                      )}
                    </div>
                    <div className="flex items-center gap-2 text-xs text-yapi-fg-subtle shrink-0">
                      {article.commitInfo.authorAvatarUrl && (
                        <img
                          src={article.commitInfo.authorAvatarUrl}
                          alt={article.commitInfo.authorName || "Author"}
                          className="w-5 h-5 rounded-full"
                        />
                      )}
                      <time dateTime={article.commitInfo.date}>
                        {new Date(article.commitInfo.date).toLocaleDateString(
                          "en-US",
                          { month: "short", day: "numeric" }
                        )}
                      </time>
                    </div>
                  </div>
                </Link>
              );
            })}
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

export const blogDataProvider = new GitHubDataProvider({
  username: "jamierpond",
  repo: "madea.blog",
  subDir: "yapi", // only the yapi folder in the repo
  token: process.env.GITHUB_TOKEN || process.env.GITHUB_PAT,
});

const SEO_CONFIG = {
  baseUrl: BASE_URL,
  siteName: "yapi",
  defaultDescription: "Updates, tutorials, and thoughts about yapi - the API development toolkit",
  authorName: "yapi",
  authorUrl: BASE_URL,
} as const;

export function createBlogConfig(): MadeaConfigWithSeo {
  return {
    dataProvider: blogDataProvider,
    username: "yapi",
    fileBrowserView: FileBrowserView,
    articleView: ArticleView,
    noRepoFoundView: NoRepoFoundView,
    landingView: LandingView,
    seo: SEO_CONFIG,
    basePath: "/blog",
  };
}

// Re-export helpers bound to config for use in pages
export async function generateBlogMetadata() {
  const config = createBlogConfig();
  const metadata = await generateMetadataForIndex(config, {
    title: "Blog | yapi",
    description: "Updates, tutorials, and thoughts about yapi",
  });
  return {
    ...metadata,
    openGraph: {
      ...metadata.openGraph,
      images: [`${OG_BASE_URL}/og/blog?title=${encodeURIComponent("Blog")}`],
    },
  };
}

export async function generateBlogArticleMetadata(slug: string[]) {
  const config = createBlogConfig();
  const metadata = await generateMetadataForArticle(config, slug);

  // Fetch article data for author and date
  const slugWithExtension = [...slug];
  slugWithExtension[slugWithExtension.length - 1] += ".md";
  const article = await blogDataProvider.getArticle(slugWithExtension.join("/"));

  const articleTitle = article?.title || slug.join("/");
  const formattedTitle = `yapi Blog | ${articleTitle}`;

  const params = new URLSearchParams({ title: articleTitle });
  if (article?.commitInfo?.authorName) params.set("author", article.commitInfo.authorName);
  if (article?.commitInfo?.date) {
    const date = new Date(article.commitInfo.date).toLocaleDateString("en-US", {
      year: "numeric", month: "short", day: "numeric",
    });
    params.set("date", date);
  }

  if (!metadata) {
    return {
      title: formattedTitle,
    }
  }
  return {
    ...metadata,
    title: formattedTitle,
    openGraph: {
      ...metadata.openGraph,
      title: formattedTitle,
      images: [`${OG_BASE_URL}/og/blog?${params.toString()}`],
    },
  };
}
