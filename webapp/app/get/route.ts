import { NextResponse, NextRequest } from 'next/server';

export async function GET(request: NextRequest) {
  return NextResponse.redirect('https://github.com/jamierpond/yapi@latest');
}
