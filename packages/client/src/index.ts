/**
 * @yapi/client - Shared CLI execution logic and types
 *
 * This package provides:
 * - Single source of truth for CLI execution (runYapi)
 * - Shared type definitions (YapiResult, YapiUIResult)
 * - Error categorization logic
 *
 * For browser-only imports (types, schemas, pure functions), use:
 *   import { ... } from '@yapi/client/types'
 *
 * This entry point includes Node-only code (child_process).
 */

import { spawn } from 'node:child_process';
import { writeFile, unlink } from 'node:fs/promises';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { randomUUID } from 'node:crypto';

// Re-export all browser-safe types and utilities
export * from './types.js';

import { YapiResultSchema, transformResultForUI } from './types.js';
import type { YapiResult, YapiUIResult } from './types.js';

// =============================================================================
// CLI Execution (Node-only)
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
 *
 * NOTE: This function is Node-only (uses child_process).
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
 *
 * NOTE: This function is Node-only (uses child_process).
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
