import type { MetadataRoute } from "next";
import { generateBlogSitemap } from "madea-blog-core";
import { blogDataProvider } from "./blog/madea.config";
import { docsDataProvider } from "./docs/madea.config";

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

  const blogEntries = await generateBlogSitemap(blogDataProvider, {
    baseUrl: BASE_URL,
    blogPath: "/blog",
  });

  const docsEntries = await generateBlogSitemap(docsDataProvider, {
    baseUrl: BASE_URL,
    blogPath: "/docs",
  });

  return [...staticPages, ...blogEntries, ...docsEntries];
}
