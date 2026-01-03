import { NextRequest, NextResponse } from "next/server";
import { generateIcon, isValidIconSize, VALID_ICON_SIZES } from "@/app/lib/generate-icon";

export const runtime = "nodejs";

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ size: string }> }
) {
  const { size: sizeParam } = await params;
  const size = parseInt(sizeParam, 10);

  if (isNaN(size) || !isValidIconSize(size)) {
    return NextResponse.json(
      { error: `Invalid size. Supported sizes: ${VALID_ICON_SIZES.join(", ")}` },
      { status: 400 }
    );
  }

  const response = await generateIcon(size);

  // Add cache headers for performance
  response.headers.set("Cache-Control", "public, max-age=31536000, immutable");

  return response;
}
