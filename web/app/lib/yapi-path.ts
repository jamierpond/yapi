import path from "path";
import { existsSync } from "fs";

export function getYapiPath(): string {
  if (process.env.VERCEL) {
    // Try multiple locations - cwd is /var/task (repo root)
    const candidates = [
      path.join(process.cwd(), "bin", "yapi"),
      path.join(process.cwd(), "web", ".next", "server", "yapi"),
      "/var/task/bin/yapi",
      "/var/task/web/.next/server/yapi",
    ];

    for (const candidate of candidates) {
      if (existsSync(candidate)) {
        return candidate;
      }
    }

    // Fallback - log what we tried
    console.error("yapi not found in:", candidates);
    return candidates[0];
  }
  return process.env.YAPI_PATH || "yapi";
}
