import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  turbopack: {

  },
  // Include the yapi binary in serverless function deployment
  outputFileTracingIncludes: {
    "/api/**/*": ["../bin/yapi"],
  },
};

export default nextConfig;
