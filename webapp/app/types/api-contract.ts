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

  /** The HTTP response body (parsed JSON or raw string) */
  responseBody: z.unknown(),

  /** HTTP status code */
  statusCode: z.number(),

  /** Request timing in milliseconds (measured server-side) */
  timing: z.number(),

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
