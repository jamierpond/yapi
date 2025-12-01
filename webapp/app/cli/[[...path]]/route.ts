import { NextRequest, NextResponse } from "next/server";

/**
 * Go vanity import handler
 *
 * Allows: go install yapi.run/cli/cmd/yapi@latest
 *
 * Go fetches this URL with ?go-get=1 and looks for a meta tag
 * that tells it where the actual repo is.
 */
export async function GET(request: NextRequest) {
  const html = `<!DOCTYPE html>
<html>
<head>
  <meta name="go-import" content="yapi.run/cli git https://github.com/jamierpond/yapi">
  <meta name="go-source" content="yapi.run/cli https://github.com/jamierpond/yapi https://github.com/jamierpond/yapi/tree/main/cli{/dir} https://github.com/jamierpond/yapi/blob/main/cli{/dir}/{file}#L{line}">
  <meta http-equiv="refresh" content="0; url=https://github.com/jamierpond/yapi/tree/main/cli">
</head>
<body>
  Redirecting to <a href="https://github.com/jamierpond/yapi/tree/main/cli">GitHub</a>...
</body>
</html>`;

  return new NextResponse(html, {
    headers: {
      "Content-Type": "text/html; charset=utf-8",
    },
  });
}
