import { NextResponse } from "next/server";
import { spawn } from "child_process";
import { z } from "zod";
import { getYapiPath } from "@/app/lib/yapi-path";
// fooooo

const VersionResponseSchema = z.object({
  version: z.string(),
  commit: z.string(),
  date: z.string(),
});

export type VersionResponse = z.infer<typeof VersionResponseSchema>;

/**
 * GET /api/yapi/version
 *
 * Returns yapi version information.
 * Uses `yapi version --json`.
 */
export async function GET() {
  try {
    const result = await new Promise<string>((resolve, reject) => {
      const proc = spawn(getYapiPath(), ["version", "--json"], {
        timeout: 5000,
      });

      let stdout = "";
      let stderr = "";

      proc.stdout.on("data", (data) => {
        stdout += data.toString();
      });

      proc.stderr.on("data", (data) => {
        stderr += data.toString();
      });

      proc.on("close", (code) => {
        if (code === 0 && stdout) {
          resolve(stdout);
        } else {
          reject(new Error(stderr || `Process exited with code ${code}`));
        }
      });

      proc.on("error", (err) => {
        reject(err);
      });
    });

    const parsed = JSON.parse(result);
    const validated = VersionResponseSchema.parse(parsed);

    return NextResponse.json(validated);
  } catch (error: unknown) {
    console.error("Error in /api/yapi/version:", error);

    return NextResponse.json(
      { error: error instanceof Error ? error.message : "Failed to get version" },
      { status: 500 }
    );
  }
}
