// Package briefing provides the embedded LLM briefing documentation.
package briefing

import _ "embed"

//go:embed LLM_BRIEFING.md
// Content holds the embedded LLM briefing documentation
var Content string
