import { ImageResponse } from "next/og";
import { readFile } from "fs/promises";
import { join } from "path";
import { COLORS } from "@/app/lib/constants";

export const runtime = "nodejs";
export const size = { width: 32, height: 32 };
export const contentType = "image/png";

export default async function Icon() {
  const jetBrainsMonoBold = await readFile(
    join(process.cwd(), "public/fonts/JetBrains_Mono/static/JetBrainsMono-Bold.ttf")
  );

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
          borderRadius: "6px",
        }}
      >
        <span
          style={{
            fontSize: "22px",
            fontFamily: "JetBrains Mono",
            fontWeight: 700,
            color: COLORS.accent,
          }}
        >
          y
        </span>
      </div>
    ),
    {
      ...size,
      fonts: [
        {
          name: "JetBrains Mono",
          data: jetBrainsMonoBold,
          style: "normal",
          weight: 700,
        },
      ],
    }
  );
}
