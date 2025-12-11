import { NextRequest, NextResponse } from 'next/server';

export async function GET(request: NextRequest) {
  const goGet = request.nextUrl.searchParams.get('go-get');

  if (goGet === '1') {
    return new NextResponse(
      `<!DOCTYPE html>
<html>
<head>
<meta name="go-import" content="yapi.run/cli git https://github.com/jamierpond/yapi">
<meta name="go-source" content="yapi.run/cli https://github.com/jamierpond/yapi https://github.com/jamierpond/yapi/tree/main{/dir} https://github.com/jamierpond/yapi/blob/main{/dir}/{file}#L{line}">
</head>
<body>go get yapi.run/cli</body>
</html>`,
      { headers: { 'Content-Type': 'text/html' } }
    );
  }

  return NextResponse.redirect('https://yapi.run');
}
