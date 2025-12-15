import { COLORS, OgContainer, YapiLogo, loadFont, createImageResponse } from "../_lib/shared";

export const dynamic = "force-dynamic";
export const fetchCache = "force-no-store";

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const title = searchParams.get("title") || "Blog";
  const author = searchParams.get("author");
  const date = searchParams.get("date");

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
          Blog
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
          marginBottom: (author || date) ? "32px" : "0",
        }}
      >
        {title}
      </div>

      {(author || date) && (
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "12px",
            fontSize: "24px",
            color: COLORS.fgMuted,
          }}
        >
          {author && <span>{author}</span>}
          {author && date && (
            <div style={{ width: "6px", height: "6px", borderRadius: "50%", backgroundColor: COLORS.fgSubtle, display: "flex" }} />
          )}
          {date && <span>{date}</span>}
        </div>
      )}
    </OgContainer>,
    fontData
  );
}
