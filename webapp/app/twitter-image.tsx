import { SITE_TITLE, OG_IMAGE_SIZE } from "@/app/lib/constants";
import { getOgImage } from "./og/route";

export const alt = SITE_TITLE;
export const size = OG_IMAGE_SIZE;
export const contentType = "image/png";

export default async function Image() {
  return getOgImage();
}
