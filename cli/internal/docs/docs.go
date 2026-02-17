// Package docs provides embedded topic-based documentation for yapi.
package docs

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/glamour"
)

//go:embed topics
var topicsFS embed.FS

// Topic represents a documentation topic.
type Topic struct {
	Name    string
	Summary string
}

// topics defines all available documentation topics.
var topics = []Topic{
	{Name: "assert", Summary: "Assertions on status, body, and headers"},
	{Name: "chain", Summary: "Multi-step request chaining"},
	{Name: "config", Summary: "YAML config field reference"},
	{Name: "environments", Summary: "Multi-environment configuration"},
	{Name: "jq", Summary: "JQ filtering and expressions"},
	{Name: "polling", Summary: "Polling with wait_for"},
	{Name: "protocols", Summary: "HTTP, gRPC, GraphQL, TCP"},
	{Name: "send", Summary: "Quick one-off requests"},
	{Name: "testing", Summary: "Test runner and CI/CD"},
	{Name: "variables", Summary: "Variable interpolation and resolution"},
}

// Get returns the content of a topic by name.
func Get(name string) (string, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	data, err := topicsFS.ReadFile("topics/" + name + ".md")
	if err != nil {
		return "", fmt.Errorf("unknown topic %q. Run 'yapi docs' to see available topics", name)
	}
	return string(data), nil
}

// List returns all available topics.
func List() []Topic {
	return topics
}

// Suggest returns the closest topic name if one is similar enough, or empty string.
func Suggest(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))
	// Check prefix matches
	var matches []string
	for _, t := range topics {
		if strings.HasPrefix(t.Name, input) {
			matches = append(matches, t.Name)
		}
	}
	if len(matches) == 1 {
		return matches[0]
	}
	// Check substring matches
	matches = matches[:0]
	for _, t := range topics {
		if strings.Contains(t.Name, input) || strings.Contains(input, t.Name) {
			matches = append(matches, t.Name)
		}
	}
	if len(matches) == 1 {
		return matches[0]
	}
	// Check summary substring
	matches = matches[:0]
	for _, t := range topics {
		if strings.Contains(strings.ToLower(t.Summary), input) {
			matches = append(matches, t.Name)
		}
	}
	if len(matches) == 1 {
		return matches[0]
	}
	return ""
}

// TopicNames returns a sorted list of topic names.
func TopicNames() []string {
	names := make([]string, len(topics))
	for i, t := range topics {
		names[i] = t.Name
	}
	sort.Strings(names)
	return names
}

// Render renders markdown content for terminal display using glamour.
func Render(markdown string) (string, error) {
	return glamour.Render(markdown, "auto")
}
