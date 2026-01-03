import { generateIcon } from "@/app/lib/generate-icon";

export const runtime = "nodejs";
export const size = { width: 256, height: 256 };
export const contentType = "image/png";

export default async function Icon() {
  return generateIcon(256);
}
