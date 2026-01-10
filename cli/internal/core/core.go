// Package core provides the main engine for executing yapi configs.
package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/executor"
	"yapi.run/cli/internal/runner"
	"yapi.run/cli/internal/validation"
)

// RequestHook is called after a request completes with stats about the execution.
// This allows the caller (main.go) to wire observability without core knowing about it.
type RequestHook func(stats map[string]any)

// Engine owns shared execution bits used by CLI, TUI, etc.
type Engine struct {
	httpClient *http.Client
	onRequest  RequestHook
}

// execFactory implements runner.ExecutorFactory using GetTransport
type execFactory struct {
	client executor.HTTPClient
}

func (f *execFactory) Create(transport string) (executor.TransportFunc, error) {
	return executor.GetTransport(transport, f.client)
}

// EngineOption configures an Engine
type EngineOption func(*Engine)

// WithRequestHook sets a hook to be called after each request
func WithRequestHook(hook RequestHook) EngineOption {
	return func(e *Engine) {
		e.onRequest = hook
	}
}

// NewEngine wires a single HTTP client.
func NewEngine(httpClient *http.Client, opts ...EngineOption) *Engine {
	e := &Engine{httpClient: httpClient}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// RunConfigResult contains the results of running a config
type RunConfigResult struct {
	Analysis  *validation.Analysis
	Result    *runner.Result
	ExpectRes *runner.ExpectationResult
	Error     error
}

// RunConfig analyzes, validates, and executes a config file.
// It never prints. Callers decide how to render diagnostics/output.
func (e *Engine) RunConfig(
	ctx context.Context,
	path string,
	opts runner.Options,
) *RunConfigResult {
	opts.ConfigFilePath = path

	// Load and analyze config
	analysis, project, err := e.loadAndAnalyze(path, opts)
	if err != nil {
		return &RunConfigResult{Error: err}
	}
	_ = project // project used only for analysis

	if analysis.HasErrors() {
		return &RunConfigResult{Analysis: analysis}
	}

	// Chain configs are returned for caller to handle
	if len(analysis.Chain) > 0 {
		return &RunConfigResult{Analysis: analysis}
	}

	// Re-expand variables if EnvOverrides is provided
	if err := e.reExpandVariables(analysis, opts); err != nil {
		return &RunConfigResult{Analysis: analysis, Error: err}
	}

	if analysis.Request == nil {
		return &RunConfigResult{Analysis: analysis}
	}

	return e.executeRequest(ctx, analysis, opts)
}

// loadAndAnalyze loads project config and analyzes the config file
func (e *Engine) loadAndAnalyze(path string, opts runner.Options) (*validation.Analysis, *config.ProjectConfigV1, error) {
	var project *config.ProjectConfigV1
	if opts.ProjectRoot != "" {
		var err error
		project, err = config.LoadProject(opts.ProjectRoot)
		if err != nil {
			if opts.ProjectEnv != "" {
				return nil, nil, fmt.Errorf("failed to load project config: %w", err)
			}
			project = nil
		}
	}

	analyzeOpts := validation.AnalyzeOptions{StrictEnv: opts.StrictEnv}
	var analysis *validation.Analysis
	var err error

	if project != nil {
		analysis, err = e.analyzeWithProject(path, project, opts, analyzeOpts)
	} else {
		analysis, err = validation.AnalyzeConfigFileWithOptions(path, analyzeOpts)
	}

	return analysis, project, err
}

// analyzeWithProject analyzes a config file with project context
func (e *Engine) analyzeWithProject(path string, project *config.ProjectConfigV1, opts runner.Options, analyzeOpts validation.AnalyzeOptions) (*validation.Analysis, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is validated user-provided config file path
	if err != nil {
		return nil, err
	}

	analyzeOpts.FilePath = path
	analyzeOpts.Project = project
	analyzeOpts.ProjectRoot = opts.ProjectRoot

	if opts.ProjectEnv != "" {
		originalDefault := project.DefaultEnvironment
		project.DefaultEnvironment = opts.ProjectEnv
		analysis, err := validation.Analyze(string(data), analyzeOpts)
		project.DefaultEnvironment = originalDefault
		return analysis, err
	}

	return validation.Analyze(string(data), analyzeOpts)
}

// reExpandVariables re-expands variables with EnvOverrides
func (e *Engine) reExpandVariables(analysis *validation.Analysis, opts runner.Options) error {
	if len(opts.EnvOverrides) == 0 || analysis.Base == nil {
		return nil
	}

	resolver := func(key string) (string, error) {
		if val, ok := os.LookupEnv(key); ok {
			return val, nil
		}
		if val, ok := opts.EnvOverrides[key]; ok {
			return val, nil
		}
		return "", nil
	}

	req, err := analysis.Base.ToDomainWithResolver(resolver)
	if err != nil {
		return err
	}
	analysis.Request = req
	return nil
}

// executeRequest runs the actual HTTP request with optional polling
func (e *Engine) executeRequest(ctx context.Context, analysis *validation.Analysis, opts runner.Options) *RunConfigResult {
	stats := ExtractConfigStats(analysis)
	start := time.Now()

	exec, err := executor.GetTransport(analysis.Request.Metadata["transport"], e.httpClient)
	if err != nil {
		return &RunConfigResult{Analysis: analysis, Error: err}
	}

	var result *runner.Result
	var runErr error

	if analysis.WaitFor != nil && len(analysis.WaitFor.Until) > 0 {
		pollResult, pollErr := runner.RunWithPolling(ctx, exec, analysis.Request, analysis.WaitFor, analysis.Warnings, opts, opts.EnvOverrides)
		if pollResult != nil {
			result = pollResult.Result
		}
		runErr = pollErr
	} else {
		result, runErr = runner.Run(ctx, exec, analysis.Request, analysis.Warnings, opts)
	}

	var expectRes *runner.ExpectationResult
	if result != nil && (analysis.Expect.Status != nil || len(analysis.Expect.Assert.Body) > 0 || len(analysis.Expect.Assert.Headers) > 0) {
		expectRes = runner.CheckExpectationsWithEnv(analysis.Expect, result, opts.EnvOverrides)
	}

	e.recordStats(stats, start, runErr, expectRes)

	if runErr != nil {
		return &RunConfigResult{Analysis: analysis, Result: result, Error: runErr}
	}

	if expectRes != nil && expectRes.Error != nil {
		return &RunConfigResult{Analysis: analysis, Result: result, ExpectRes: expectRes, Error: expectRes.Error}
	}

	return &RunConfigResult{Analysis: analysis, Result: result, ExpectRes: expectRes}
}

// recordStats records request stats if a hook is configured
func (e *Engine) recordStats(stats map[string]any, start time.Time, runErr error, expectRes *runner.ExpectationResult) {
	if e.onRequest == nil {
		return
	}
	stats["duration_ms"] = time.Since(start).Milliseconds()
	stats["success"] = runErr == nil && (expectRes == nil || expectRes.Error == nil)
	if runErr != nil {
		stats["error_type"] = "execution"
	} else if expectRes != nil && expectRes.Error != nil {
		stats["error_type"] = "assertion_failed"
	}
	e.onRequest(stats)
}

// RunChain executes a chain configuration
func (e *Engine) RunChain(
	ctx context.Context,
	base *config.ConfigV1,
	chain []config.ChainStep,
	opts runner.Options,
	analysis *validation.Analysis,
) (*runner.ChainResult, error) {
	stats := ExtractConfigStats(analysis)
	start := time.Now()

	result, err := runner.RunChain(ctx, &execFactory{e.httpClient}, base, chain, opts)

	if e.onRequest != nil {
		stats["duration_ms"] = time.Since(start).Milliseconds()
		stats["success"] = err == nil
		if err != nil {
			stats["error_type"] = "chain_execution"
		}
		e.onRequest(stats)
	}

	return result, err
}
