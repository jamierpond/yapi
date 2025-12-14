import { COLORS, OgContainer, YapiLogo, loadFont, createImageResponse } from "../_lib/shared";

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const title = searchParams.get("title") || "Documentation";

  const fontData = await loadFont();
  const fontSize = title.length > 50 ? "42px" : title.length > 30 ? "52px" : "64px";

  return createImageResponse(
    <OgContainer>
      <div style={{ display: "flex", marginBottom: "24px" }}>
        <YapiLogo size="small" />
      </div>

      <div
        style={{
          display: "flex",
          padding: "8px 20px",
          backgroundColor: COLORS.accent,
          borderRadius: "6px",
          marginBottom: "32px",
        }}
      >
        <span style={{ fontSize: "24px", color: "#fff", textTransform: "uppercase", letterSpacing: "2px" }}>
          Docs
        </span>
      </div>

      <div
        style={{
          display: "flex",
          fontSize,
          fontWeight: "bold",
          color: COLORS.fg,
          textAlign: "center",
          lineHeight: 1.2,
          maxWidth: "90%",
        }}
      >
        {title}
      </div>
    </OgContainer>,
    fontData
  );
}
