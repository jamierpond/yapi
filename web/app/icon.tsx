import { ImageResponse } from "next/og";
import { COLORS } from "@/app/lib/constants";

export const runtime = "nodejs";
export const size = { width: 32, height: 32 };
export const contentType = "image/png";

export default function Icon() {
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
