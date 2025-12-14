import { renderMadeaBlogPage } from "madea-blog-core";
import { createDocsConfig } from "./madea.config";

export const dynamic = "force-static";

const CONFIG = createDocsConfig();

export default async function DocsPage() {
  return renderMadeaBlogPage(CONFIG);
}
