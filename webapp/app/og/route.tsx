import { ImageResponse } from "next/og";
import { SITE_TITLE, COLORS, OG_IMAGE_SIZE } from "@/app/lib/constants";

export async function getOgImage() {
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
          fontFamily: "monospace",
        }}
      >
        {/* Background gradient accents - matching Landing page */}
        <div
          style={{
            position: "absolute",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            background: `linear-gradient(135deg, ${COLORS.accent}1a 0%, transparent 50%)`,
          }}
        />
        <div
          style={{
            position: "absolute",
            top: "25%",
            right: "25%",
            width: "400px",
            height: "400px",
            background: `radial-gradient(circle, ${COLORS.accent}0d 0%, transparent 70%)`,
            borderRadius: "50%",
          }}
        />

        {/* Vignette effect */}
        <div
          style={{
            position: "absolute",
            width: "100%",
            height: "100%",
            background:
              "radial-gradient(circle at 50% 50%, transparent 30%, rgba(0, 0, 0, 0.4) 100%)",
          }}
        />

        {/* Sheep emoji */}
        <div
          style={{
            fontSize: "120px",
            marginBottom: "20px",
            display: "flex",
          }}
        >
          🐑
        </div>

        {/* Main title */}
        <div
          style={{
            display: "flex",
            flexDirection: "row",
            alignItems: "center",
            justifyContent: "center",
            marginBottom: "30px",
          }}
        >
          <div
            style={{
              fontSize: "160px",
              fontWeight: "bold",
              color: COLORS.fg,
              fontFamily: "monospace",
              letterSpacing: "-4px",
              display: "flex",
            }}
          >
            {SITE_TITLE}
          </div>
        </div>

        {/* Tagline */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "12px",
            padding: "16px 32px",
            backgroundColor: COLORS.bgElevated,
            border: `1px solid ${COLORS.border}`,
            borderRadius: "9999px",
            marginBottom: "40px",
          }}
        >
          <div
            style={{
              width: "10px",
              height: "10px",
              borderRadius: "50%",
              backgroundColor: COLORS.accent,
              display: "flex",
            }}
          />
          <span
            style={{
              fontSize: "24px",
              fontFamily: "monospace",
              color: COLORS.fgMuted,
              textTransform: "uppercase",
              letterSpacing: "2px",
            }}
          >
            Bash-powered YAML API workbench
          </span>
        </div>

        {/* Subtitle */}
        <div
          style={{
            fontSize: "48px",
            fontWeight: "bold",
            color: COLORS.fg,
            fontFamily: "monospace",
            textAlign: "center",
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
          }}
        >
          <span>API testing</span>
          <span style={{ color: COLORS.accent }}>in YAML</span>
        </div>

        {/* Protocol badges */}
        <div
          style={{
            display: "flex",
            gap: "20px",
            marginTop: "50px",
          }}
        >
          {["HTTP", "gRPC", "TCP"].map((protocol) => (
            <div
              key={protocol}
              style={{
                padding: "12px 28px",
                backgroundColor: COLORS.bgElevated,
                border: `1px solid ${COLORS.border}`,
                borderRadius: "12px",
                fontSize: "22px",
                fontFamily: "monospace",
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
    }
  );
}

export async function GET() {
  return getOgImage();
}
