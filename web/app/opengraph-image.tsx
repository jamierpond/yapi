import { SITE_TITLE, OG_IMAGE_SIZE } from "@/app/lib/constants";
import { getHomepageOgImage } from "./og/_lib/shared";

export const alt = SITE_TITLE;
export const size = OG_IMAGE_SIZE;
export const contentType = "image/png";

export default async function Image() {
  return getHomepageOgImage();
}
