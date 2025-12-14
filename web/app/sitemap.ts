import type { MetadataRoute } from "next";
import { generateBlogSitemap } from "madea-blog-core";
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

  // Generate blog sitemap entries
  const dataProvider = new LocalFsDataProvider({
    contentDir: path.join(process.cwd(), "app/blog/_content"),
    authorName: "yapi",
  });

  const blogEntries = await generateBlogSitemap(dataProvider, {
    baseUrl: BASE_URL,
    blogPath: "/blog",
  });

  return [...staticPages, ...blogEntries];
}
