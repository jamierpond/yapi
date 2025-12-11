import { NextRequest, NextResponse } from "next/server";

/**
 * Middleware to handle Go vanity imports
 *
 * When Go fetches a URL with ?go-get=1, it expects a meta tag
 * that tells it where the actual repo is.
 *
 * This allows: go install yapi.run/cmd/yapi@latest
 */
export function middleware(request: NextRequest) {
  const { searchParams } = request.nextUrl;

  // Only intercept go-get requests
  if (searchParams.get("go-get") === "1") {
    const html = `<!DOCTYPE html>
<html>
<head>
  <meta name="go-import" content="yapi.run git https://github.com/jamierpond/yapi">
  <meta name="go-source" content="yapi.run https://github.com/jamierpond/yapi https://github.com/jamierpond/yapi/tree/main{/dir} https://github.com/jamierpond/yapi/blob/main{/dir}/{file}#L{line}">
</head>
<body>
  go get yapi.run
</body>
</html>`;

    return new NextResponse(html, {
      headers: {
        "Content-Type": "text/html; charset=utf-8",
      },
    });
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    // Match root and all subpaths for go module paths
    "/",
    "/cmd/:path*",
    "/internal/:path*",
  ],
};
