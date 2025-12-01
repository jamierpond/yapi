import { NextResponse } from 'next/server';
export async function GET() {
  return NextResponse.redirect("https://github.com/jamierpond/yapi");
}
