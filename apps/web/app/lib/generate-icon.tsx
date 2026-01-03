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

  // Scale font size proportionally
  const fontSize = Math.round(size * 0.6);
  // Scale border radius proportionally
  const borderRadius = Math.round(size * 0.22);

  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: `linear-gradient(145deg, #151515 0%, ${COLORS.bg} 50%, #050505 100%)`,
          borderRadius: `${borderRadius}px`,
          position: "relative",
          overflow: "hidden",
        }}
      >
        {/* Primary glow - bottom left orange */}
        <div
          style={{
            position: "absolute",
            bottom: "-30%",
            left: "-30%",
            width: "100%",
            height: "100%",
            background: `radial-gradient(circle at center, ${COLORS.accent}55 0%, ${COLORS.accent}22 30%, transparent 60%)`,
          }}
        />
        {/* Secondary glow - top right subtle purple */}
        <div
          style={{
            position: "absolute",
            top: "-40%",
            right: "-40%",
            width: "90%",
            height: "90%",
            background: "radial-gradient(circle at center, #6366f133 0%, transparent 50%)",
          }}
        />
        {/* Center highlight */}
        <div
          style={{
            position: "absolute",
            top: "10%",
            left: "10%",
            width: "80%",
            height: "80%",
            background: `radial-gradient(circle at 30% 30%, ${COLORS.accent}18 0%, transparent 50%)`,
          }}
        />
        <span
          style={{
            fontSize: `${fontSize}px`,
            fontFamily: "JetBrains Mono",
            fontWeight: 700,
            color: COLORS.accent,
            position: "relative",
            textShadow: `0 0 ${Math.round(size * 0.15)}px ${COLORS.accent}66`,
            marginTop: `${Math.round(size * -0.08)}px`,
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
