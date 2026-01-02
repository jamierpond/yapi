import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  turbopack: {

  },
  // Include the yapi binary in serverless function deployment
  // Path relative to apps/web -> monorepo root bin/yapi
  outputFileTracingIncludes: {
    "/api/**/*": ["../../bin/yapi"],
  },
};

export default nextConfig;
