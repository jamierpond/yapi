import { NextRequest, NextResponse } from "next/server";
import { exec } from "child_process";
import { promisify } from "util";
import { writeFile, unlink } from "fs/promises";
import { tmpdir } from "os";
import { join } from "path";
import { parse } from "yaml";
import {
  ExecuteRequestSchema,
  ExecuteSuccessResponseSchema,
  ExecuteErrorResponseSchema,
} from "@yapi/ui";
import { getYapiPath } from "@/app/lib/yapi-path";

const execAsync = promisify(exec);

// SSRF Protection: Define blocked IP ranges
const IS_IP_V4 = /^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$/;
const PRIVATE_IP_RANGES = [
  /^127\./,           // Localhost
  /^10\./,            // Local LAN
  /^192\.168\./,      // Local LAN
  /^172\.(1[6-9]|2[0-9]|3[0-1])\./, // Docker/Local LAN
  /^169\.254\./,      // Cloud Metadata (AWS/GCP/Azure)
  /^0\.0\.0\.0/       // All interfaces
];

// Helper to validate URL for SSRF protection
function isSafeUrl(urlStr: string): boolean {
  try {
    const url = new URL(urlStr);

    // Block non-http/grpc protocols (e.g., file://)
    if (!['http:', 'https:', 'grpc:', 'grpcs:', 'tcp:'].includes(url.protocol)) {
      return false;
    }

    const hostname = url.hostname;

    // Block "localhost" explicitly
    if (hostname === 'localhost') return false;

    // Check against Private IP regex
    if (IS_IP_V4.test(hostname)) {
      if (PRIVATE_IP_RANGES.some(regex => regex.test(hostname))) {
        return false;
      }
    }

    // NOTE: This does NOT prevent DNS Rebinding (where attacker maps google.com -> 127.0.0.1)
    // To fix that requires a custom DNS resolver, which is not a "quick fix".

    return true;
  } catch (e) {
    return false; // Invalid URL
  }
}

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

    // SSRF Protection: Validate URL(s) in YAML
    try {
      const parsed = parse(yaml);

      // Collect all URLs to validate - either from top-level url or from chain
      const urlsToValidate: string[] = [];

      if (parsed.url) {
        urlsToValidate.push(parsed.url);
      }

      if (Array.isArray(parsed.chain)) {
        for (const step of parsed.chain) {
          if (step.url) {
            urlsToValidate.push(step.url);
          }
        }
      }

      if (urlsToValidate.length === 0) {
        const errorResponse = ExecuteErrorResponseSchema.parse({
          success: false,
          error: "YAML must contain a 'url' field or a 'chain' with URLs",
          errorType: "VALIDATION_ERROR",
        });
        return NextResponse.json(errorResponse, { status: 400 });
      }

      // Check all URLs for SSRF
      for (const url of urlsToValidate) {
        if (!isSafeUrl(url)) {
          const errorResponse = ExecuteErrorResponseSchema.parse({
            success: false,
            error: `Security Violation: Access to local/private networks is blocked for URL: ${url}`,
            errorType: "SSRF_BLOCKED",
          });
          return NextResponse.json(errorResponse, { status: 403 });
        }
      }
    } catch (e) {
      const errorResponse = ExecuteErrorResponseSchema.parse({
        success: false,
        error: "Invalid YAML",
        errorType: "YAML_PARSE_ERROR",
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

    // Execute yapi command with --json flag for structured output
    const startTime = Date.now();
    const { stdout, stderr } = await execAsync(`${getYapiPath()} run --json "${tempFile}"`, {
      timeout: 30000, // 30 second timeout
      maxBuffer: 10 * 1024 * 1024, // 10MB buffer
    });
    const timing = Date.now() - startTime;

    // Parse JSON output from CLI
    let cliOutput: any;
    try {
      cliOutput = JSON.parse(stdout);
    } catch (e) {
      // Fallback: if JSON parsing fails, treat as raw output
      const response = ExecuteSuccessResponseSchema.parse({
        success: true,
        responseBody: stdout,
        timing,
      });
      return NextResponse.json(response);
    }

    // The body from CLI is a string - try to parse it as JSON if possible
    // If it's already valid JSON, parse it so the frontend can display it nicely
    // If it's not JSON, keep it as a string
    let responseBody: unknown;
    if (typeof cliOutput.body === 'string' && cliOutput.body.trim().length > 0) {
      try {
        responseBody = JSON.parse(cliOutput.body);
      } catch {
        // Not JSON, keep as string
        responseBody = cliOutput.body;
      }
    } else {
      responseBody = cliOutput.body;
    }

    // Build comprehensive success response from CLI output
    const response = ExecuteSuccessResponseSchema.parse({
      success: cliOutput.success !== false,
      responseBody: responseBody,
      transport: cliOutput.transport,
      statusCode: cliOutput.statusCode,
      timing: cliOutput.timing || timing,
      headers: cliOutput.headers,
      requestUrl: cliOutput.requestUrl,
      method: cliOutput.method,
      service: cliOutput.service,
      contentType: cliOutput.contentType,
      sizeBytes: cliOutput.sizeBytes,
      sizeLines: cliOutput.sizeLines,
      sizeChars: cliOutput.sizeChars,
      warnings: cliOutput.warnings,
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
