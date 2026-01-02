import { renderMadeaBlogPage } from "madea-blog-core";
import { createBlogConfig, generateBlogMetadata } from "./madea.config";

export const revalidate = 600; // Cache for 10 minutes

export const generateMetadata = generateBlogMetadata;

const CONFIG = createBlogConfig();

export default async function BlogPage() {
  return renderMadeaBlogPage(CONFIG);
}
