import { NextRequest, NextResponse } from "next/server";
import { parse } from "yaml";
import { runYapiForUI, categorizeError } from "@yapi/client";
import { ExecuteRequestSchema } from "@yapi/ui";
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

function isSafeUrl(urlStr: string): boolean {
  try {
    const url = new URL(urlStr);
    if (!['http:', 'https:', 'grpc:', 'grpcs:', 'tcp:'].includes(url.protocol)) {
      return false;
    }
    const hostname = url.hostname;
    if (hostname === 'localhost') return false;
    if (IS_IP_V4.test(hostname)) {
      if (PRIVATE_IP_RANGES.some(regex => regex.test(hostname))) {
        return false;
      }
    }
    return true;
  } catch {
    return false;
  }
}

/**
 * POST /api/execute
 */
export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const parseResult = ExecuteRequestSchema.safeParse(body);

    if (!parseResult.success) {
      return NextResponse.json({
        success: false,
        error: "Invalid request format",
        errorType: "VALIDATION_ERROR",
        details: parseResult.error.format(),
      }, { status: 400 });
    }

    const { yaml } = parseResult.data;

    if (!yaml || yaml.trim().length === 0) {
      return NextResponse.json({
        success: false,
        error: "YAML content is empty",
        errorType: "VALIDATION_ERROR",
      }, { status: 400 });
    }

    // SSRF Protection
    try {
      const parsed = parse(yaml);
      const urlsToValidate: string[] = [];

      if (parsed.url) urlsToValidate.push(parsed.url);
      if (Array.isArray(parsed.chain)) {
        for (const step of parsed.chain) {
          if (step.url) urlsToValidate.push(step.url);
        }
      }

      if (urlsToValidate.length === 0) {
        return NextResponse.json({
          success: false,
          error: "YAML must contain a 'url' field or a 'chain' with URLs",
          errorType: "VALIDATION_ERROR",
        }, { status: 400 });
      }

      for (const url of urlsToValidate) {
        if (!isSafeUrl(url)) {
          return NextResponse.json({
            success: false,
            error: `Security Violation: Access to local/private networks is blocked for URL: ${url}`,
            errorType: "SSRF_BLOCKED",
          }, { status: 403 });
        }
      }
    } catch {
      return NextResponse.json({
        success: false,
        error: "Invalid YAML",
        errorType: "YAML_PARSE_ERROR",
      }, { status: 400 });
    }

    // Execute and return UI-ready result
    const result = await runYapiForUI({
      executablePath: getYapiPath(),
      input: { type: 'content', yaml },
      timeout: 30000,
    });

    return NextResponse.json(result);

  } catch (error: unknown) {
    console.error("Error in /api/execute:", error);
    const errorMessage = error instanceof Error ? error.message : "An unexpected error occurred";

    return NextResponse.json({
      success: false,
      error: errorMessage,
      errorType: categorizeError(errorMessage),
    }, { status: 500 });
  }
}
