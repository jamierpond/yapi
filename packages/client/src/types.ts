/**
 * @yapi/client/types - Browser-safe types and utilities
 *
 * This file contains ONLY browser-safe exports:
 * - Type definitions
 * - Zod schemas
 * - Pure functions (no Node APIs)
 *
 * IMPORTANT: Keep YapiResultSchema in sync with Go CLI output (cmd/yapi/main.go).
 */

import { z } from 'zod';

// =============================================================================
// Shared Type Definitions
// =============================================================================

/**
 * Error types for categorizing failures.
 * Used by both web and extension UIs.
 */
export const ErrorType = z.enum([
  'YAML_PARSE_ERROR',
  'VALIDATION_ERROR',
  'NETWORK_ERROR',
  'SSRF_BLOCKED',
  'TIMEOUT',
  'UNKNOWN',
]);
export type ErrorType = z.infer<typeof ErrorType>;

/**
 * Schema for individual assertion result.
 * Mirrors jsonAssertionResult struct in cmd/yapi/main.go.
 */
export const AssertionResultSchema = z.object({
  expression: z.string(),
  passed: z.boolean(),
  actual: z.string().optional(),
  expected: z.string().optional(),
  leftSide: z.string().optional(),
  operator: z.string().optional(),
  error: z.string().optional(),
});

export type AssertionResult = z.infer<typeof AssertionResultSchema>;

/**
 * Schema for assertion summary with results.
 */
export const AssertionsSchema = z.object({
  total: z.number(),
  passed: z.number(),
  results: z.array(AssertionResultSchema).optional(),
});

export type Assertions = z.infer<typeof AssertionsSchema>;

/**
 * Schema for CLI JSON output.
 * Mirrors the jsonOutput struct in cmd/yapi/main.go.
 */
export const YapiResultSchema = z.object({
  success: z.boolean(),
  body: z.string(),
  transport: z.string().optional(),
  statusCode: z.number().optional(),
  headers: z.record(z.string()).optional(),
  requestUrl: z.string().optional(),
  method: z.string().optional(),
  service: z.string().optional(),
  contentType: z.string().optional(),
  sizeBytes: z.number().optional(),
  sizeLines: z.number().optional(),
  sizeChars: z.number().optional(),
  timing: z.number(),
  warnings: z.array(z.string()).optional(),
  error: z.string().optional(),
  assertions: AssertionsSchema.optional(),
});

export type YapiResult = z.infer<typeof YapiResultSchema>;

/**
 * Success response formatted for UI consumption.
 */
export interface YapiUISuccess {
  success: true;
  responseBody: unknown;
  transport?: string;
  statusCode?: number;
  timing: number;
  headers?: Record<string, string>;
  requestUrl?: string;
  method?: string;
  service?: string;
  contentType?: string;
  sizeBytes?: number;
  sizeLines?: number;
  sizeChars?: number;
  warnings?: string[];
  assertions?: Assertions;
}

/**
 * Error response formatted for UI consumption.
 */
export interface YapiUIError {
  success: false;
  error: string;
  errorType: ErrorType;
  details?: unknown;
}

/**
 * Union type for UI consumption.
 */
export type YapiUIResult = YapiUISuccess | YapiUIError;

// =============================================================================
// Error Categorization
// =============================================================================

/**
 * Categorize an error message to determine its type.
 * Used by both web and extension to provide consistent error feedback.
 */
export function categorizeError(errorMessage: string): ErrorType {
  const lowerMsg = errorMessage.toLowerCase();
  if (lowerMsg.includes('timeout')) return 'TIMEOUT';
  if (lowerMsg.includes('yaml') || lowerMsg.includes('parse')) return 'YAML_PARSE_ERROR';
  if (lowerMsg.includes('validation') || lowerMsg.includes('invalid')) return 'VALIDATION_ERROR';
  if (lowerMsg.includes('network') || lowerMsg.includes('connection') || lowerMsg.includes('econnrefused')) return 'NETWORK_ERROR';
  if (lowerMsg.includes('ssrf') || lowerMsg.includes('blocked')) return 'SSRF_BLOCKED';
  return 'UNKNOWN';
}

// =============================================================================
// Result Transformation
// =============================================================================

/**
 * Transform raw CLI result to UI-friendly format.
 * Handles JSON parsing of body and error categorization.
 */
export function transformResultForUI(result: YapiResult): YapiUIResult {
  if (!result.success) {
    return {
      success: false,
      error: result.error || 'Unknown error',
      errorType: categorizeError(result.error || ''),
      details: result.body || undefined,
    };
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

  return {
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
    assertions: result.assertions,
  };
}
