import { ImageResponse } from "next/og";
import { COLORS, OG_IMAGE_SIZE, OG_ALT } from "@/app/lib/constants";
import fs from "fs/promises";
import path from "path";

export { COLORS, OG_IMAGE_SIZE, OG_ALT };

export async function loadFont() {
  return fs.readFile(
    path.join(process.cwd(), "public", "fonts", "JetBrains_Mono", "static", "JetBrainsMono-Bold.ttf")
  );
}

export function createImageResponse(element: React.ReactElement, fontData: Buffer) {
  return new ImageResponse(element, {
    ...OG_IMAGE_SIZE,
    fonts: [{ name: "JetBrains Mono", data: fontData, style: "normal", weight: 700 }],
  });
}

export function OgBackground() {
  return (
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
  );
}

export function YapiLogo({ size = "large" }: { size?: "large" | "small" }) {
  const fontSize = size === "large" ? "180px" : "72px";
  const letterSpacing = size === "large" ? "-6px" : "-3px";

  return (
    <div style={{ display: "flex", flexDirection: "row", alignItems: "baseline", justifyContent: "center" }}>
      <span style={{ fontSize, fontWeight: "bold", color: COLORS.fg, letterSpacing }}>y</span>
      <span style={{ fontSize, fontWeight: "bold", color: COLORS.accent, letterSpacing }}>a</span>
      <span style={{ fontSize, fontWeight: "bold", color: COLORS.fg, letterSpacing }}>pi</span>
    </div>
  );
}

export function OgContainer({ children }: { children: React.ReactNode }) {
  return (
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
        padding: "60px",
      }}
    >
      <OgBackground />
      {children}
    </div>
  );
}

export async function getHomepageOgImage() {
  const fontData = await loadFont();

  return createImageResponse(
    <OgContainer>
      <div style={{ display: "flex", flexDirection: "column", alignItems: "center", marginBottom: "40px" }}>
        <YapiLogo size="large" />
      </div>

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
        <div style={{ width: "14px", height: "14px", borderRadius: "50%", backgroundColor: COLORS.accent, display: "flex" }} />
        <span style={{ fontSize: "36px", color: COLORS.fgMuted, textTransform: "uppercase", letterSpacing: "3px" }}>
          Offline-first YAML API client
        </span>
      </div>

      <div style={{ display: "flex", gap: "24px" }}>
        {["HTTP", "gRPC", "GraphQL", "TCP"].map((p) => (
          <div
            key={p}
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
            {p}
          </div>
        ))}
      </div>
    </OgContainer>,
    fontData
  );
}
