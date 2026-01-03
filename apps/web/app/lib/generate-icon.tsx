import { ImageResponse } from "next/og";
import { readFile } from "fs/promises";
import { join } from "path";
import { COLORS } from "@/app/lib/constants";

const VALID_SIZES = [16, 32, 64, 128, 256] as const;
export type IconSize = (typeof VALID_SIZES)[number];

export function isValidIconSize(size: number): size is IconSize {
  return VALID_SIZES.includes(size as IconSize);
}

export const VALID_ICON_SIZES = VALID_SIZES;

async function loadFont(): Promise<Buffer> {
  return readFile(
    join(process.cwd(), "public/fonts/JetBrains_Mono/static/JetBrainsMono-Bold.ttf")
  );
}

async function loadSheepImage(): Promise<string> {
  const sheepBuffer = await readFile(join(process.cwd(), "public/sheep.png"));
  return `data:image/png;base64,${sheepBuffer.toString("base64")}`;
}

export async function generateIcon(size: IconSize): Promise<ImageResponse> {
  const fontData = await loadFont();
  const sheepDataUrl = await loadSheepImage();

  const borderRadius = Math.round(size * 0.22);
  const sheepSize = Math.round(size * 0.85);

  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: COLORS.bg,
          borderRadius: `${borderRadius}px`,
        }}
      >
        <img
          src={sheepDataUrl}
          width={sheepSize}
          height={sheepSize}
          style={{ objectFit: "contain" }}
        />
      </div>
    ),
    {
      width: size,
      height: size,
      fonts: [
        {
          name: "JetBrains Mono",
          data: fontData,
          style: "normal",
          weight: 700,
        },
      ],
    }
  );
}
