package core

import (
	"context"
	"net/http"

	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/executor"
	"yapi.run/cli/internal/runner"
	"yapi.run/cli/internal/validation"
)

// Engine owns shared execution bits used by CLI, TUI, etc.
type Engine struct {
	Factory *executor.Factory
}

// NewEngine wires a single HTTP client and executor factory.
func NewEngine(httpClient *http.Client) *Engine {
	return &Engine{Factory: executor.NewFactory(httpClient)}
}

// RunConfig analyzes, validates, and executes a config file.
// It never prints. Callers decide how to render diagnostics/output.
func (e *Engine) RunConfig(
	ctx context.Context,
	path string,
	opts runner.Options,
) (*validation.Analysis, *runner.Result, error) {
	analysis, err := validation.AnalyzeConfigFile(path)
	if err != nil {
		return nil, nil, err
	}

	if analysis.HasErrors() {
		return analysis, nil, nil
	}

	// Check if this is a chain config
	if len(analysis.Chain) > 0 {
		// For chains, return analysis only - caller handles execution
		return analysis, nil, nil
	}

	if analysis.Request == nil {
		return analysis, nil, nil
	}

	exec, err := e.Factory.Create(analysis.Request.Metadata["transport"])
	if err != nil {
		return analysis, nil, err
	}

	result, err := runner.Run(ctx, exec, analysis.Request, analysis.Warnings, opts)
	return analysis, result, err
}

// RunChain executes a chain configuration
func (e *Engine) RunChain(
	ctx context.Context,
	chain []config.ChainStep,
	opts runner.Options,
) (*runner.ChainResult, error) {
	return runner.RunChain(ctx, e.Factory, chain, opts)
}
