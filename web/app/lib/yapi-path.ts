import path from "path";

export function getYapiPath(): string {
  if (process.env.VERCEL) {
    return path.join(process.cwd(), ".next", "server", "yapi");
  }
  return process.env.YAPI_PATH || "yapi";
}
