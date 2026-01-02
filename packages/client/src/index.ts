/**
 * @yapi/client - Shared CLI execution logic
 *
 * This package provides a single source of truth for spawning the yapi CLI,
 * handling temp files, and parsing output. Used by both the web API route
 * and the VS Code extension.
 *
 * IMPORTANT: Keep in sync with Go CLI output format (cmd/yapi/main.go).
 * If you change the JSON output structure in Go, update YapiResultSchema here.
 */

import { spawn } from 'node:child_process';
import { writeFile, unlink } from 'node:fs/promises';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { randomUUID } from 'node:crypto';
import { z } from 'zod';

/**
 * Schema for CLI JSON output.
 * Mirrors the jsonOutput struct in cmd/yapi/main.go.
 */
const YapiResultSchema = z.object({
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
      // Don't await cleanup, just fire and forget to keep response fast
      unlink(targetFilePath).catch(() => {});
    }
  }
}

/**
 * Internal helper to spawn process with promise wrapper and timeout
 */
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

    // Timeout safety
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

/**
 * Parses raw CLI output into structured result
 */
function parseYapiOutput({ stdout, stderr }: { stdout: string; stderr: string }): YapiResult {
  try {
    // 1. Try strict JSON parse
    const raw = JSON.parse(stdout);

    // 2. Validate against schema
    // Use safeParse to avoid throwing if the CLI changes slightly,
    // preferring to return a partial result than crashing.
    const parsed = YapiResultSchema.safeParse(raw);

    if (parsed.success) {
      return parsed.data;
    } else {
      console.error('Yapi schema validation failed:', parsed.error);
      // Fallback for valid JSON but invalid schema
      return {
        success: false,
        body: stdout,
        timing: 0,
        error: 'Invalid JSON structure returned from yapi CLI',
        warnings: [parsed.error.message]
      };
    }
  } catch {
    // 3. Fallback: CLI crashed or returned non-JSON text
    return {
      success: false,
      body: stdout || stderr, // If stdout is empty, show stderr
      timing: 0,
      error: stderr || 'Failed to parse JSON output',
    };
  }
}
