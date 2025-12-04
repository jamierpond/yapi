import { NextRequest, NextResponse } from "next/server";
import { spawn } from "child_process";
import {
  ValidateRequestSchema,
  ValidateResponseSchema,
  type ValidateResponse,
} from "@/app/types/api-contract";

/**
 * POST /api/validate
 *
 * Validates yapi YAML and returns diagnostics.
 * Uses `yapi validate --json -` with stdin.
 */
export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const parseResult = ValidateRequestSchema.safeParse(body);

    if (!parseResult.success) {
      const errorResponse: ValidateResponse = {
        valid: false,
        diagnostics: [{
          severity: "error",
          message: "Invalid request format",
          line: 0,
          col: 0,
        }],
        warnings: [],
      };
      return NextResponse.json(errorResponse, { status: 400 });
    }

    const { yaml } = parseResult.data;

    // Spawn yapi validate and pipe yaml to stdin
    const result = await new Promise<string>((resolve, reject) => {
      const proc = spawn("yapi", ["validate", "--json", "-"], {
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
        // yapi validate returns exit code 1 for validation errors, but still outputs valid JSON
        if (stdout) {
          resolve(stdout);
        } else {
          reject(new Error(stderr || `Process exited with code ${code}`));
        }
      });

      proc.on("error", (err) => {
        reject(err);
      });

      // Write yaml to stdin and close
      proc.stdin.write(yaml);
      proc.stdin.end();
    });

    // Parse the JSON output from yapi
    const parsed = JSON.parse(result);
    const validated = ValidateResponseSchema.parse(parsed);

    return NextResponse.json(validated);
  } catch (error: unknown) {
    console.error("Error in /api/validate:", error);

    const errorResponse: ValidateResponse = {
      valid: false,
      diagnostics: [{
        severity: "error",
        message: error instanceof Error ? error.message : "Validation failed",
        line: 0,
        col: 0,
      }],
      warnings: [],
    };

    return NextResponse.json(errorResponse, { status: 500 });
  }
}
