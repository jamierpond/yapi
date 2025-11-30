import { NextRequest, NextResponse } from "next/server";
import { exec } from "child_process";
import { promisify } from "util";
import { writeFile, unlink } from "fs/promises";
import { tmpdir } from "os";
import { join } from "path";
import {
  ExecuteRequestSchema,
  ExecuteSuccessResponseSchema,
  ExecuteErrorResponseSchema,
} from "@/app/types/api-contract";

const execAsync = promisify(exec);

/**
 * POST /api/execute
 *
 * Executes a yapi YAML request and returns the response.
 */
export async function POST(request: NextRequest) {
  let tempFile: string | null = null;

  try {
    // Parse and validate request body
    const body = await request.json();
    const parseResult = ExecuteRequestSchema.safeParse(body);

    if (!parseResult.success) {
      const errorResponse = ExecuteErrorResponseSchema.parse({
        success: false,
        error: "Invalid request format",
        errorType: "VALIDATION_ERROR",
        details: parseResult.error.format(),
      });
      return NextResponse.json(errorResponse, { status: 400 });
    }

    const { yaml } = parseResult.data;

    // Validate that we have content
    if (!yaml || yaml.trim().length === 0) {
      const errorResponse = ExecuteErrorResponseSchema.parse({
        success: false,
        error: "YAML content is empty",
        errorType: "VALIDATION_ERROR",
      });
      return NextResponse.json(errorResponse, { status: 400 });
    }

    // Write YAML to temporary file
    const timestamp = Date.now();
    const randomId = Math.random().toString(36).substring(7);
    tempFile = join(tmpdir(), `yapi-${timestamp}-${randomId}.yaml`);
    await writeFile(tempFile, yaml, "utf-8");

    console.log("Executing yapi with file:", tempFile);
    console.log("YAML content:", yaml);

    // Execute yapi command with timing
    const startTime = Date.now();
    const { stdout, stderr } = await execAsync(`yapi -c "${tempFile}"`, {
      timeout: 30000, // 30 second timeout
      maxBuffer: 10 * 1024 * 1024, // 10MB buffer
    });
    const timing = Date.now() - startTime;

    // Parse yapi output
    let responseBody: unknown;
    let statusCode = 200;

    try {
      // Try to parse as JSON
      responseBody = JSON.parse(stdout);
    } catch {
      // If not JSON, return as raw text
      responseBody = stdout;
    }

    // Build success response
    const response = ExecuteSuccessResponseSchema.parse({
      success: true,
      responseBody,
      statusCode,
      timing,
    });

    return NextResponse.json(response);
  } catch (error: any) {
    console.error("Error in /api/execute:", error);

    let errorType: "YAML_PARSE_ERROR" | "VALIDATION_ERROR" | "NETWORK_ERROR" | "SSRF_BLOCKED" | "TIMEOUT" | "UNKNOWN" = "UNKNOWN";
    let errorMessage = "An unexpected error occurred";

    if (error.killed || error.signal === "SIGTERM") {
      errorType = "TIMEOUT";
      errorMessage = "Request timed out after 30 seconds";
    } else if (error.code === "ENOENT") {
      errorType = "UNKNOWN";
      errorMessage = "yapi command not found in PATH";
    } else if (error.stderr) {
      errorMessage = error.stderr;
      // Try to categorize based on stderr content
      if (error.stderr.includes("YAML") || error.stderr.includes("parse")) {
        errorType = "YAML_PARSE_ERROR";
      } else if (error.stderr.includes("validation") || error.stderr.includes("invalid")) {
        errorType = "VALIDATION_ERROR";
      } else if (error.stderr.includes("network") || error.stderr.includes("connection")) {
        errorType = "NETWORK_ERROR";
      }
    } else if (error instanceof Error) {
      errorMessage = error.message;
    }

    const errorResponse = ExecuteErrorResponseSchema.parse({
      success: false,
      error: errorMessage,
      errorType,
      details: error.stderr || error.stdout || undefined,
    });

    return NextResponse.json(errorResponse, { status: 500 });
  } finally {
    // Clean up temp file
    if (tempFile) {
      try {
        await unlink(tempFile);
      } catch {
        // Ignore cleanup errors
      }
    }
  }
}
