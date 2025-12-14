import type { MetadataRoute } from "next";
import { generateBlogSitemap } from "madea-blog-core";
import { GitHubDataProvider } from "madea-blog-core/providers/github";
import { LocalFsDataProvider } from "madea-blog-core/providers/local-fs";
import path from "path";

const BASE_URL = "https://yapi.run";

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const staticPages: MetadataRoute.Sitemap = [
    {
      url: BASE_URL,
      lastModified: new Date(),
      changeFrequency: "weekly",
      priority: 1,
    },
    {
      url: `${BASE_URL}/playground`,
      lastModified: new Date(),
      changeFrequency: "weekly",
      priority: 0.8,
    },
  ];

  // Generate blog sitemap entries (from GitHub)
  const blogProvider = new GitHubDataProvider({
    username: "jamierpond",
    repo: "madea.blog",
    subDir: "yapi",
  });

  const blogEntries = await generateBlogSitemap(blogProvider, {
    baseUrl: BASE_URL,
    blogPath: "/blog",
  });

  // Generate docs sitemap entries (from local filesystem)
  const docsProvider = new LocalFsDataProvider({
    contentDir: path.join(process.cwd(), "app/_docs"),
    authorName: "yapi",
    sourceUrl: "https://github.com/jamierpond/yapi",
  });

  const docsEntries = await generateBlogSitemap(docsProvider, {
    baseUrl: BASE_URL,
    blogPath: "/docs",
  });

  return [...staticPages, ...blogEntries, ...docsEntries];
}
