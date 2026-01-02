import { renderMadeaBlogPage } from "madea-blog-core";
import { createBlogConfig, generateBlogArticleMetadata } from "../madea.config";

export const revalidate = 600; // Cache for 10 minutes

const CONFIG = createBlogConfig();

interface PageProps {
  params: Promise<{ slug: string[] }>;
}

export async function generateMetadata({ params }: PageProps) {
  const { slug } = await params;
  return generateBlogArticleMetadata(slug);
}

export default async function ArticlePage({ params }: PageProps) {
  const resolvedParams = await params;
  // Add .md extension back for the data provider
  const slugWithExtension = [...resolvedParams.slug];
  slugWithExtension[slugWithExtension.length - 1] += ".md";

  return renderMadeaBlogPage(CONFIG, Promise.resolve({ slug: slugWithExtension }));
}
