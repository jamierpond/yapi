import { ImageResponse } from "next/og";
import { COLORS, OG_IMAGE_SIZE } from "@/app/lib/constants";
import fs from "fs/promises";
import path from "path";

export async function getOgImage() {
  const fontData = await fs.readFile(
    path.join(process.cwd(), "public", "fonts", "JetBrains_Mono", "static", "JetBrainsMono-Bold.ttf")
  );

  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          backgroundColor: COLORS.bg,
          position: "relative",
          fontFamily: "JetBrains Mono",
        }}
      >
        {/* Top-left radial glow bloom */}
        <div
          style={{
            position: "absolute",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: `radial-gradient(ellipse 120% 100% at 0% 0%, ${COLORS.accent}28 0%, transparent 50%)`,
          }}
        />

        {/* Main title with accent "a" */}
        <div
          style={{
            display: "flex",
            flexDirection: "row",
            alignItems: "baseline",
            justifyContent: "center",
            marginBottom: "40px",
          }}
        >
          <span
            style={{
              fontSize: "180px",
              fontWeight: "bold",
              color: COLORS.fg,
              letterSpacing: "-6px",
            }}
          >
            y
          </span>
          <span
            style={{
              fontSize: "180px",
              fontWeight: "bold",
              color: COLORS.accent,
              letterSpacing: "-6px",
            }}
          >
            a
          </span>
          <span
            style={{
              fontSize: "180px",
              fontWeight: "bold",
              color: COLORS.fg,
              letterSpacing: "-6px",
            }}
          >
            pi
          </span>
        </div>

        {/* Tagline */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "16px",
            padding: "20px 40px",
            backgroundColor: COLORS.bgElevated,
            border: `1px solid ${COLORS.border}`,
            borderRadius: "9999px",
            marginBottom: "50px",
          }}
        >
          <div
            style={{
              width: "14px",
              height: "14px",
              borderRadius: "50%",
              backgroundColor: COLORS.accent,
              display: "flex",
            }}
          />
          <span
            style={{
              fontSize: "36px",
              color: COLORS.fgMuted,
              textTransform: "uppercase",
              letterSpacing: "3px",
            }}
          >
            Offline-first YAML API client
          </span>
        </div>

        {/* Protocol badges */}
        <div
          style={{
            display: "flex",
            gap: "24px",
          }}
        >
          {["HTTP", "gRPC", "GraphQL", "TCP"].map((protocol) => (
            <div
              key={protocol}
              style={{
                padding: "18px 40px",
                backgroundColor: COLORS.bgElevated,
                border: `1px solid ${COLORS.border}`,
                borderRadius: "12px",
                fontSize: "36px",
                color: COLORS.fgMuted,
                display: "flex",
              }}
            >
              {protocol}
            </div>
          ))}
        </div>
      </div>
    ),
    {
      ...OG_IMAGE_SIZE,
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

export async function GET() {
  return getOgImage();
}
