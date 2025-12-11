import { ImageResponse } from "next/og";
import { COLORS } from "@/app/lib/constants";

export const runtime = "nodejs";
export const size = { width: 180, height: 180 };
export const contentType = "image/png";

export default function AppleIcon() {
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
          borderRadius: "36px",
        }}
      >
        <span
          style={{
            fontSize: "120px",
            fontWeight: "bold",
            color: COLORS.accent,
          }}
        >
          y
        </span>
      </div>
    ),
    { ...size }
  );
}
