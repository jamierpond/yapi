import { NextRequest, NextResponse } from "next/server";
import { parse } from "yaml";
import { runYapi, type YapiResult } from "@yapi/client";
import {
  ExecuteRequestSchema,
  ExecuteSuccessResponseSchema,
  ExecuteErrorResponseSchema,
} from "@yapi/ui";
import { getYapiPath } from "@/app/lib/yapi-path";

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
  } catch {
    return false; // Invalid URL
  }
}

/**
 * Transform CLI YapiResult to UI ExecuteResponse format
 */
function transformResult(result: YapiResult) {
  if (!result.success) {
    return ExecuteErrorResponseSchema.parse({
      success: false,
      error: result.error || 'Unknown error',
      errorType: categorizeError(result.error || ''),
      details: result.body || undefined,
    });
  }

  // Try to parse body as JSON for nicer display
  let responseBody: unknown;
  if (typeof result.body === 'string' && result.body.trim().length > 0) {
    try {
      responseBody = JSON.parse(result.body);
    } catch {
      responseBody = result.body;
    }
  } else {
    responseBody = result.body;
  }

  return ExecuteSuccessResponseSchema.parse({
    success: true,
    responseBody,
    transport: result.transport,
    statusCode: result.statusCode,
    timing: result.timing,
    headers: result.headers,
    requestUrl: result.requestUrl,
    method: result.method,
    service: result.service,
    contentType: result.contentType,
    sizeBytes: result.sizeBytes,
    sizeLines: result.sizeLines,
    sizeChars: result.sizeChars,
    warnings: result.warnings,
  });
}

/**
 * Categorize error message to determine error type
 */
function categorizeError(errorMessage: string): "YAML_PARSE_ERROR" | "VALIDATION_ERROR" | "NETWORK_ERROR" | "SSRF_BLOCKED" | "TIMEOUT" | "UNKNOWN" {
  const lowerMsg = errorMessage.toLowerCase();
  if (lowerMsg.includes('timeout')) return 'TIMEOUT';
  if (lowerMsg.includes('yaml') || lowerMsg.includes('parse')) return 'YAML_PARSE_ERROR';
  if (lowerMsg.includes('validation') || lowerMsg.includes('invalid')) return 'VALIDATION_ERROR';
  if (lowerMsg.includes('network') || lowerMsg.includes('connection')) return 'NETWORK_ERROR';
  return 'UNKNOWN';
}

/**
 * POST /api/execute
 *
 * Executes a yapi YAML request and returns the response.
 */
export async function POST(request: NextRequest) {
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
    } catch {
      const errorResponse = ExecuteErrorResponseSchema.parse({
        success: false,
        error: "Invalid YAML",
        errorType: "YAML_PARSE_ERROR",
      });
      return NextResponse.json(errorResponse, { status: 400 });
    }

    // Execute yapi using shared client
    const result = await runYapi({
      executablePath: getYapiPath(),
      input: { type: 'content', yaml },
      timeout: 30000,
    });

    // Transform to UI format
    const response = transformResult(result);
    return NextResponse.json(response);

  } catch (error: unknown) {
    console.error("Error in /api/execute:", error);

    const errorMessage = error instanceof Error ? error.message : "An unexpected error occurred";

    const errorResponse = ExecuteErrorResponseSchema.parse({
      success: false,
      error: errorMessage,
      errorType: categorizeError(errorMessage),
    });

    return NextResponse.json(errorResponse, { status: 500 });
  }
}
