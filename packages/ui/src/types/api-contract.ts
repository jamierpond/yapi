/**
 * FE/BE API Contract
 *
 * This file re-exports types from @yapi/client and adds request schemas.
 * The response types are defined in @yapi/client to keep a single source of truth.
 *
 * Type hierarchy:
 * - Go CLI (cmd/yapi/main.go) -> defines jsonOutput struct
 * - @yapi/client -> defines YapiResultSchema (mirrors Go), YapiUIResult (for UI)
 * - @yapi/ui -> re-exports and adds request validation
 */

import { z } from "zod";

// Re-export types from @yapi/client/types (browser-safe, no Node APIs)
export {
  type YapiResult,
  type YapiUIResult,
  type YapiUISuccess,
  type YapiUIError,
  type ErrorType,
  type Assertions,
  type AssertionResult,
  YapiResultSchema,
  AssertionResultSchema,
  AssertionsSchema,
  ErrorType as ErrorTypeEnum,
  categorizeError,
  transformResultForUI,
} from "@yapi/client/types";

// =============================================================================
// Request Schema: POST /api/execute
// =============================================================================

/**
 * The request payload sent from the editor to execute a yapi request
 */
export const ExecuteRequestSchema = z.object({
  /** The raw YAML string from the editor */
  yaml: z.string(),
});

export type ExecuteRequest = z.infer<typeof ExecuteRequestSchema>;

// =============================================================================
// Response Types (re-exported from @yapi/client)
// =============================================================================

// For backwards compatibility, alias the types
import type { YapiUISuccess, YapiUIError, YapiUIResult } from "@yapi/client/types";

export type ExecuteSuccessResponse = YapiUISuccess;
export type ExecuteErrorResponse = YapiUIError;
export type ExecuteResponse = YapiUIResult;

// =============================================================================
// Helper Functions
// =============================================================================

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

// =============================================================================
// Validation API: POST /api/validate
// =============================================================================

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
