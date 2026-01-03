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

export async function generateIcon(size: IconSize): Promise<ImageResponse> {
  const fontData = await loadFont();

  const fontSize = Math.round(size * 0.55);
  const borderRadius = Math.round(size * 0.2);
  const gridSize = Math.max(4, Math.round(size * 0.1));

  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: `linear-gradient(135deg, #2d1a0a 0%, #1a0f05 30%, ${COLORS.bg} 70%, #08080c 100%)`,
          borderRadius: `${borderRadius}px`,
          position: "relative",
          overflow: "hidden",
        }}
      >
        {/* Grid pattern */}
        <div
          style={{
            position: "absolute",
            inset: 0,
            backgroundImage: `linear-gradient(to right, ${COLORS.accent}15 1px, transparent 1px), linear-gradient(to bottom, ${COLORS.accent}15 1px, transparent 1px)`,
            backgroundSize: `${gridSize}px ${gridSize}px`,
          }}
        />
        {/* Center glow */}
        <div
          style={{
            position: "absolute",
            inset: 0,
            background: `radial-gradient(circle at center, ${COLORS.accent}22 0%, transparent 70%)`,
          }}
        />
        <span
          style={{
            fontSize: `${fontSize}px`,
            fontFamily: "JetBrains Mono",
            fontWeight: 700,
            color: COLORS.accent,
            marginTop: `${Math.round(size * -0.1)}px`,
            position: "relative",
          }}
        >
          y
        </span>
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
