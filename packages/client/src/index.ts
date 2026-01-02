/**
 * @yapi/client - Shared CLI execution logic and types
 *
 * This package provides:
 * - Single source of truth for CLI execution (runYapi)
 * - Shared type definitions (YapiResult, YapiUIResult)
 * - Error categorization logic
 *
 * IMPORTANT: Keep YapiResultSchema in sync with Go CLI output (cmd/yapi/main.go).
 * If you change the JSON output structure in Go, update this file.
 */

import { spawn } from 'node:child_process';
import { writeFile, unlink } from 'node:fs/promises';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { randomUUID } from 'node:crypto';
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
  };
}

// =============================================================================
// CLI Execution
// =============================================================================

export interface YapiOptions {
  /** Path to the yapi executable */
  executablePath: string;
  /** The content of the yapi file OR a path to an existing file */
  input: { type: 'content'; yaml: string } | { type: 'file'; path: string };
  /** Environment to use (passed to --env) */
  env?: string;
  /** Execution timeout in ms (default: 30000) */
  timeout?: number;
}

/**
 * Executes the yapi CLI and returns structured output.
 * Handles both "file on disk" (VS Code) and "raw content" (Web Playground) scenarios.
 */
export async function runYapi(options: YapiOptions): Promise<YapiResult> {
  const { executablePath, input, env, timeout = 30000 } = options;

  let targetFilePath: string;
  let isTempFile = false;

  // 1. Prepare Input File
  if (input.type === 'file') {
    targetFilePath = input.path;
  } else {
    // Write raw content to temp file
    isTempFile = true;
    const tempDir = tmpdir();
    const fileName = `yapi-exec-${randomUUID()}.yapi.yml`;
    targetFilePath = join(tempDir, fileName);
    await writeFile(targetFilePath, input.yaml, 'utf8');
  }

  try {
    // 2. Build Arguments
    const args = ['run', '--json', targetFilePath];
    if (env) {
      args.push('--env', env);
    }

    // 3. Execute
    const output = await spawnYapiProcess(executablePath, args, timeout);

    // 4. Parse Output
    return parseYapiOutput(output);

  } catch (err: unknown) {
    // Handle system/spawn errors (not API errors, which are handled in JSON)
    const message = err instanceof Error ? err.message : String(err);
    return {
      success: false,
      body: '',
      timing: 0,
      error: `Internal Execution Error: ${message}`,
      warnings: [],
    };
  } finally {
    // 5. Cleanup
    if (isTempFile) {
      unlink(targetFilePath).catch(() => {});
    }
  }
}

/**
 * Execute yapi and return UI-ready result.
 * Convenience function that combines runYapi + transformResultForUI.
 */
export async function runYapiForUI(options: YapiOptions): Promise<YapiUIResult> {
  const result = await runYapi(options);
  return transformResultForUI(result);
}

// =============================================================================
// Internal Helpers
// =============================================================================

function spawnYapiProcess(
  cmd: string,
  args: string[],
  timeoutMs: number
): Promise<{ stdout: string; stderr: string }> {
  return new Promise((resolve, reject) => {
    const child = spawn(cmd, args, {
      env: { ...process.env },
    });

    let stdout = '';
    let stderr = '';
    let completed = false;

    const timer = setTimeout(() => {
      if (!completed) {
        completed = true;
        child.kill();
        reject(new Error(`Execution timed out after ${timeoutMs}ms`));
      }
    }, timeoutMs);

    child.stdout.on('data', (data) => {
      stdout += data.toString();
    });

    child.stderr.on('data', (data) => {
      stderr += data.toString();
    });

    child.on('error', (err) => {
      if (!completed) {
        completed = true;
        clearTimeout(timer);
        reject(err);
      }
    });

    child.on('close', () => {
      if (!completed) {
        completed = true;
        clearTimeout(timer);
        resolve({ stdout, stderr });
      }
    });
  });
}

function parseYapiOutput({ stdout, stderr }: { stdout: string; stderr: string }): YapiResult {
  try {
    const raw = JSON.parse(stdout);
    const parsed = YapiResultSchema.safeParse(raw);

    if (parsed.success) {
      return parsed.data;
    } else {
      console.error('Yapi schema validation failed:', parsed.error);
      return {
        success: false,
        body: stdout,
        timing: 0,
        error: 'Invalid JSON structure returned from yapi CLI',
        warnings: [parsed.error.message]
      };
    }
  } catch {
    return {
      success: false,
      body: stdout || stderr,
      timing: 0,
      error: stderr || 'Failed to parse JSON output',
    };
  }
}
