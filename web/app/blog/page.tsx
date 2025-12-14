import { renderMadeaBlogPage } from "madea-blog-core";
import { createBlogConfig } from "./madea.config";

export const dynamic = "force-dynamic";

const CONFIG = createBlogConfig();

export default async function BlogPage() {
  return renderMadeaBlogPage(CONFIG);
}
