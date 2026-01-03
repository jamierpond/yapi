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

  // Scale font size proportionally (roughly 68% of icon size works well)
  const fontSize = Math.round(size * 0.68);
  // Scale border radius proportionally (roughly 18% of icon size)
  const borderRadius = Math.round(size * 0.18);

  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          backgroundColor: COLORS.bg,
          borderRadius: `${borderRadius}px`,
          position: "relative",
          overflow: "hidden",
        }}
      >
        {/* Gradient glow effect inspired by landing page */}
        <div
          style={{
            position: "absolute",
            top: "-50%",
            left: "-50%",
            width: "200%",
            height: "200%",
            background: `radial-gradient(circle at 50% 50%, ${COLORS.accent}33 0%, transparent 50%)`,
          }}
        />
        <span
          style={{
            fontSize: `${fontSize}px`,
            fontFamily: "JetBrains Mono",
            fontWeight: 700,
            color: COLORS.accent,
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
