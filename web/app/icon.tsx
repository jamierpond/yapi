import { ImageResponse } from "next/og";
import { COLORS } from "@/app/lib/constants";

export const runtime = "nodejs";
export const size = { width: 32, height: 32 };
export const contentType = "image/png";

export default async function Icon() {
  const jetBrainsMonoBold = await fetch(
    new URL("https://fonts.gstatic.com/s/jetbrainsmono/v18/tDbY2o-flEEny0FZhsfKu5WU4zr3E_BX0PnT8RD8yKxjPVmUsaaDhw.woff2")
  ).then((res) => res.arrayBuffer());

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
