import { COLORS, OgContainer, YapiLogo, loadFont, createImageResponse } from "../_lib/shared";

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const title = searchParams.get("title") || "Blog";
  const author = searchParams.get("author");
  const date = searchParams.get("date");

  const fontData = await loadFont();
  const fontSize = title.length > 50 ? "42px" : title.length > 30 ? "52px" : "64px";

  return createImageResponse(
    <OgContainer>
      <div style={{ display: "flex", marginBottom: "32px" }}>
        <YapiLogo size="small" />
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
          marginBottom: "40px",
        }}
      >
        {title}
      </div>

      {(author || date) && (
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "16px",
            padding: "16px 32px",
            backgroundColor: COLORS.bgElevated,
            border: `1px solid ${COLORS.border}`,
            borderRadius: "9999px",
          }}
        >
          {author && (
            <span style={{ fontSize: "28px", color: COLORS.fgMuted }}>{author}</span>
          )}
          {author && date && (
            <div style={{ width: "6px", height: "6px", borderRadius: "50%", backgroundColor: COLORS.border, display: "flex" }} />
          )}
          {date && (
            <span style={{ fontSize: "28px", color: COLORS.fgMuted }}>{date}</span>
          )}
        </div>
      )}
    </OgContainer>,
    fontData
  );
}
