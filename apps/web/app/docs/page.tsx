import { renderMadeaBlogPage } from "madea-blog-core";
import { createDocsConfig, generateDocsMetadata } from "./madea.config";

export const dynamic = "force-static";

export const generateMetadata = generateDocsMetadata;

const CONFIG = createDocsConfig();

export default async function DocsPage() {
  return renderMadeaBlogPage(CONFIG);
}
