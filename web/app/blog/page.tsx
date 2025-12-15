import { renderMadeaBlogPage } from "madea-blog-core";
import { createBlogConfig, generateBlogMetadata } from "./madea.config";

export const dynamic = "force-dynamic";
export const fetchCache = "force-no-store";

export const generateMetadata = generateBlogMetadata;

const CONFIG = createBlogConfig();

export default async function BlogPage() {
  return renderMadeaBlogPage(CONFIG);
}
