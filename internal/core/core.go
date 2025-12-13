// Package core provides the main engine for executing yapi configs.
package core

import (
	"context"
	"net/http"
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
	factory   *executor.Factory
	onRequest RequestHook
}

// EngineOption configures an Engine
type EngineOption func(*Engine)

// WithRequestHook sets a hook to be called after each request
func WithRequestHook(hook RequestHook) EngineOption {
	return func(e *Engine) {
		e.onRequest = hook
	}
}

// NewEngine wires a single HTTP client and executor factory.
func NewEngine(httpClient *http.Client, opts ...EngineOption) *Engine {
	e := &Engine{factory: executor.NewFactory(httpClient)}
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
	analysis, err := validation.AnalyzeConfigFile(path)
	if err != nil {
		return &RunConfigResult{Error: err}
	}

	if analysis.HasErrors() {
		return &RunConfigResult{Analysis: analysis}
	}

	// Check if this is a chain config
	if len(analysis.Chain) > 0 {
		// For chains, return analysis only - caller handles execution
		return &RunConfigResult{Analysis: analysis}
	}

	if analysis.Request == nil {
		return &RunConfigResult{Analysis: analysis}
	}

	// Extract config stats for hook
	stats := ExtractConfigStats(analysis)
	start := time.Now()

	exec, err := e.factory.Create(analysis.Request.Metadata["transport"])
	if err != nil {
		return &RunConfigResult{Analysis: analysis, Error: err}
	}

	result, runErr := runner.Run(ctx, exec, analysis.Request, analysis.Warnings, opts)

	// Check expectations if present
	var expectRes *runner.ExpectationResult
	if result != nil && (analysis.Expect.Status != nil || len(analysis.Expect.Assert) > 0) {
		expectRes = runner.CheckExpectations(analysis.Expect, result)
	}

	// Call hook with request stats (if configured)
	if e.onRequest != nil {
		stats["duration_ms"] = time.Since(start).Milliseconds()
		stats["success"] = runErr == nil && (expectRes == nil || expectRes.Error == nil)
		if runErr != nil {
			stats["error_type"] = "execution"
		} else if expectRes != nil && expectRes.Error != nil {
			stats["error_type"] = "assertion_failed"
		}
		e.onRequest(stats)
	}

	if runErr != nil {
		return &RunConfigResult{Analysis: analysis, Result: result, Error: runErr}
	}

	if expectRes != nil && expectRes.Error != nil {
		return &RunConfigResult{Analysis: analysis, Result: result, ExpectRes: expectRes, Error: expectRes.Error}
	}

	return &RunConfigResult{Analysis: analysis, Result: result, ExpectRes: expectRes}
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

	result, err := runner.RunChain(ctx, e.factory, base, chain, opts)

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
