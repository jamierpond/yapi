import { NextRequest, NextResponse } from "next/server";
import { exec } from "child_process";
import { promisify } from "util";
import {
  ValidateRequestSchema,
  ValidateResponseSchema,
  type ValidateResponse,
} from "@/app/types/api-contract";

const execAsync = promisify(exec);

/**
 * POST /api/validate
 *
 * Validates yapi YAML and returns diagnostics.
 * Uses `yapi validate --json` via stdin.
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

    // Run yapi validate --json with yaml piped to stdin
    const { stdout } = await execAsync(
      `echo ${JSON.stringify(yaml)} | yapi validate --json -`,
      {
        timeout: 5000,
        maxBuffer: 1024 * 1024,
        shell: "/bin/bash",
      }
    );

    // Parse the JSON output from yapi
    const result = JSON.parse(stdout);
    const validated = ValidateResponseSchema.parse(result);

    return NextResponse.json(validated);
  } catch (error: unknown) {
    console.error("Error in /api/validate:", error);

    // Even on error, try to return a valid response structure
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
