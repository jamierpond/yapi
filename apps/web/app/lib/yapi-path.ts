import path from "path";
import { existsSync } from "fs";

export function getYapiPath(): string {
  // Development: use local env or system command
  if (process.env.NODE_ENV === "development") {
    return process.env.YAPI_PATH || "yapi";
  }

  // Production (Vercel): binary traced to root/bin folder
  const vercelPath = path.join(process.cwd(), "bin", "yapi");
  if (existsSync(vercelPath)) {
    return vercelPath;
  }

  // Fallback for monorepo root vs app root differences
  const altPath = path.join(process.cwd(), "..", "bin", "yapi");
  if (existsSync(altPath)) {
    return altPath;
  }

  throw new Error(`yapi binary not found. Searched: ${vercelPath}, ${altPath}`);
}
