// Package briefing provides the embedded LLM briefing documentation.
package briefing

import _ "embed"

// Content holds the embedded LLM briefing documentation
//
//go:embed LLM_BRIEFING.md
var Content string
