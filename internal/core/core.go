package core

import (
	"context"
	"net/http"

	"yapi.run/cli/internal/executor"
	"yapi.run/cli/internal/runner"
	"yapi.run/cli/internal/validation"
)

// Engine owns shared execution bits used by CLI, TUI, etc.
type Engine struct {
	factory *executor.Factory
}

// NewEngine wires a single HTTP client and executor factory.
func NewEngine(httpClient *http.Client) *Engine {
	return &Engine{factory: executor.NewFactory(httpClient)}
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

	if analysis.HasErrors() || analysis.Request == nil {
		return analysis, nil, nil
	}

	exec, err := e.factory.Create(analysis.Request.Metadata["transport"])
	if err != nil {
		return analysis, nil, err
	}

	result, err := runner.Run(ctx, exec, analysis.Request, analysis.Warnings, opts)
	return analysis, result, err
}
