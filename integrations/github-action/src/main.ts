import * as core from '@actions/core';
import * as exec from '@actions/exec';
import { spawn } from 'child_process';

async function run(): Promise<void> {
  try {
    // -------------------------------------------------------------------------
    // 1. PARSE INPUTS
    // -------------------------------------------------------------------------
    const startCmds = core.getMultilineInput('start');
    const command = core.getInput('command') || 'yapi test .';
    const skipInstall = core.getBooleanInput('skip-install');

    // Get version from the action ref (e.g., 'v0.5.0' or 'main')
    const actionRef = process.env.GITHUB_ACTION_REF || 'main';
    const version = actionRef === 'main' ? 'latest' : actionRef;

    // -------------------------------------------------------------------------
    // 2. INSTALL YAPI
    // -------------------------------------------------------------------------
    if (skipInstall) {
      core.startGroup('Verifying pre-installed Yapi');
      try {
        await exec.exec('yapi', ['version']);
        core.info('Using local/pre-installed yapi');
      } catch {
        core.setFailed('skip-install is enabled but yapi is not found in PATH. Please install yapi before running this action.');
        process.exit(1);
      }
      core.endGroup();
    } else {
      // Check if yapi is already installed
      let yapiInstalled = false;
      try {
        await exec.exec('yapi', ['version'], { silent: true });
        yapiInstalled = true;
      } catch {
        yapiInstalled = false;
      }

      if (yapiInstalled) {
        core.startGroup('Using pre-installed Yapi');
        await exec.exec('yapi', ['version']);
        core.endGroup();
      } else {
        core.startGroup(`Installing Yapi ${version}`);

        // Use the unified install script that works across platforms
        let installCmd = 'curl -fsSL https://yapi.run/install/linux.sh | bash';

        // If a specific version is requested (not 'latest'), set YAPI_VERSION env var
        if (version !== 'latest') {
          core.info(`Installing yapi version: ${version}`);
          installCmd = `curl -fsSL https://yapi.run/install/linux.sh | YAPI_VERSION=${version} bash`;
        }

        // Use sh -c to properly handle the pipe operator
        await exec.exec('sh', ['-c', installCmd]);

        // Add yapi to PATH for the rest of this step
        const yapiPath = `${process.env.HOME}/.yapi/bin`;
        core.addPath(yapiPath);

        core.endGroup();
      }
    }

    // -------------------------------------------------------------------------
    // 3. START BACKGROUND SERVERS
    // -------------------------------------------------------------------------
    if (startCmds.length > 0) {
      core.startGroup('Starting background services');

      for (const cmd of startCmds) {
        if (!cmd.trim()) continue; // Skip empty lines

        core.info(`> ${cmd}`);

        // We use 'spawn' instead of @actions/exec because we don't want to await
        // the process. We want it to run in the background.
        // 'shell: true' allows piping and using '&&' in the command string.
        const subprocess = spawn(cmd, {
          detached: true,
          stdio: 'inherit', // Pipe logs to the GitHub Action console
          shell: true,
        });

        // We don't 'unref()' here because we want the logs to keep streaming.
        // GitHub Actions runner will automatically kill this process tree
        // when the step finishes.
        if (!subprocess.pid) {
          throw new Error(`Failed to spawn command: ${cmd}`);
        }
      }
      core.endGroup();
    }

    // -------------------------------------------------------------------------
    // 4. RUN YAPI TESTS
    // -------------------------------------------------------------------------
    // Note: Use yapi's native --wait-on flag for health checks, e.g.:
    //   yapi test . --wait-on=http://localhost:3000/healthz --wait-timeout=60s
    core.startGroup('Running Yapi Tests');
    // We use @actions/exec here because we WANT to await this and fail if it fails
    const exitCode = await exec.exec(command);
    core.endGroup();

    if (exitCode !== 0) {
      core.setFailed(`Yapi tests failed with exit code ${exitCode}`);
      process.exit(1);
    }

  } catch (error) {
    if (error instanceof Error) {
      core.setFailed(error.message);
    } else {
      core.setFailed('An unexpected error occurred');
    }
    process.exit(1);
  }
}

run().then(() => {
  process.exit(0);
}).catch(() => {
  process.exit(1);
});
