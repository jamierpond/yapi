import { z } from "zod";

/**
 * FE/BE API Contract
 *
 * This file defines the contract between the frontend and backend
 * for the yapi playground. All API interactions should conform to these schemas.
 */

// ============================================================================
// Request Schema: POST /api/execute
// ============================================================================

/**
 * The request payload sent from the editor to execute a yapi request
 */
export const ExecuteRequestSchema = z.object({
  /** The raw YAML string from the editor */
  yaml: z.string(),
});

export type ExecuteRequest = z.infer<typeof ExecuteRequestSchema>;

// ============================================================================
// Response Schema: POST /api/execute
// ============================================================================

/**
 * Successful execution response
 */
export const ExecuteSuccessResponseSchema = z.object({
  /** Whether the execution was successful */
  success: z.literal(true),

  /** The response body (parsed JSON or raw string) */
  responseBody: z.unknown(),

  /** Transport type used (http, grpc, tcp, graphql) */
  transport: z.enum(["http", "grpc", "tcp", "graphql"]).optional(),

  /** Status code (HTTP only, 0 or omitted for gRPC/TCP) */
  statusCode: z.number().optional(),

  /** Request timing in milliseconds */
  timing: z.number(),

  /** Response metadata (headers for HTTP/GraphQL, can be empty for gRPC/TCP) */
  headers: z.record(z.string(), z.string()).optional(),

  /** The full request URL/endpoint that was executed */
  requestUrl: z.string().optional(),

  /** Method/RPC name (GET/POST for HTTP, RPC method for gRPC, "tcp" for TCP) */
  method: z.string().optional(),

  /** Service name (gRPC only) */
  service: z.string().optional(),

  /** Content-Type of the response */
  contentType: z.string().optional(),

  /** Response size in bytes */
  sizeBytes: z.number().optional(),

  /** Number of lines in response body */
  sizeLines: z.number().optional(),

  /** Number of characters in response body */
  sizeChars: z.number().optional(),

  /** Warnings generated during execution */
  warnings: z.array(z.string()).optional(),

  /** Optional: The parsed YAML config (for debugging) */
  parsedConfig: z.unknown().optional(),
});

/**
 * Error response when execution fails
 */
export const ExecuteErrorResponseSchema = z.object({
  /** Whether the execution was successful */
  success: z.literal(false),

  /** Error message */
  error: z.string(),

  /** Error type for categorization */
  errorType: z.enum([
    "YAML_PARSE_ERROR",
    "VALIDATION_ERROR",
    "NETWORK_ERROR",
    "SSRF_BLOCKED",
    "TIMEOUT",
    "UNKNOWN"
  ]),

  /** Optional: Additional error details for debugging */
  details: z.unknown().optional(),
});

/**
 * Union of success and error responses
 */
export const ExecuteResponseSchema = z.union([
  ExecuteSuccessResponseSchema,
  ExecuteErrorResponseSchema,
]);

export type ExecuteSuccessResponse = z.infer<typeof ExecuteSuccessResponseSchema>;
export type ExecuteErrorResponse = z.infer<typeof ExecuteErrorResponseSchema>;
export type ExecuteResponse = z.infer<typeof ExecuteResponseSchema>;

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Type guard to check if response is successful
 */
export function isSuccessResponse(
  response: ExecuteResponse
): response is ExecuteSuccessResponse {
  return response.success === true;
}

/**
 * Type guard to check if response is an error
 */
export function isErrorResponse(
  response: ExecuteResponse
): response is ExecuteErrorResponse {
  return response.success === false;
}

// ============================================================================
// Validation API: POST /api/validate
// ============================================================================

/**
 * Request payload for validation
 */
export const ValidateRequestSchema = z.object({
  yaml: z.string(),
});

export type ValidateRequest = z.infer<typeof ValidateRequestSchema>;

/**
 * A single diagnostic from the validator
 */
export const DiagnosticSchema = z.object({
  severity: z.enum(["error", "warning", "info"]),
  field: z.string().optional(),
  message: z.string(),
  line: z.number(),
  col: z.number(),
});

export type Diagnostic = z.infer<typeof DiagnosticSchema>;

/**
 * Validation response
 */
export const ValidateResponseSchema = z.object({
  valid: z.boolean(),
  diagnostics: z.array(DiagnosticSchema),
  warnings: z.array(z.string()),
});

export type ValidateResponse = z.infer<typeof ValidateResponseSchema>;
