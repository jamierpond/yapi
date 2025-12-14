import { renderMadeaBlogPage } from "madea-blog-core";
import { createBlogConfig } from "../madea.config";

export const dynamic = "force-dynamic";

const CONFIG = createBlogConfig();

interface PageProps {
  params: Promise<{ slug: string[] }>;
}

export default async function ArticlePage({ params }: PageProps) {
  const resolvedParams = await params;
  // Add .md extension back for the data provider
  const slugWithExtension = [...resolvedParams.slug];
  slugWithExtension[slugWithExtension.length - 1] += ".md";

  return renderMadeaBlogPage(CONFIG, Promise.resolve({ slug: slugWithExtension }));
}
