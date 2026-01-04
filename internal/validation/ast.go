package validation

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Location represents a file position (protocol-agnostic)
type Location struct {
	File string // Absolute file path
	Line int    // 0-indexed line number
	Col  int    // 0-indexed column number
}

// FindVarPositionInYAML finds the position of a variable in a YAML config file.
// section is a path like ["environments", "dev", "vars"] to navigate to.
// varName is the key to find within that section.
func FindVarPositionInYAML(projectRoot string, varName string, section []string) (*Location, error) {
	// Try both .yml and .yaml extensions
	var configPath string
	ymlPath := filepath.Join(projectRoot, "yapi.config.yml")
	yamlPath := filepath.Join(projectRoot, "yapi.config.yaml")

	if _, err := os.Stat(ymlPath); err == nil {
		configPath = ymlPath
	} else if _, err := os.Stat(yamlPath); err == nil {
		configPath = yamlPath
	} else {
		return nil, fmt.Errorf("config file not found: tried %s and %s in %s",
			filepath.Base(ymlPath), filepath.Base(yamlPath), projectRoot)
	}

	contentBytes, err := os.ReadFile(configPath) // #nosec G304 -- configPath is constructed from validated projectRoot
	if err != nil {
		return nil, err
	}

	var root yaml.Node
	if err := yaml.Unmarshal(contentBytes, &root); err != nil {
		return nil, err
	}

	// Navigate to the section (e.g., ["environments", "dev", "vars"])
	currentNode := &root
	if len(root.Content) > 0 {
		currentNode = root.Content[0] // Get the document content
	}

	for _, key := range section {
		valueNode := FindNodeInMapping(currentNode, key)
		if valueNode == nil {
			return nil, fmt.Errorf("section not found: %s", key)
		}
		currentNode = valueNode
	}

	// Now find the key node for the variable
	keyNode := FindKeyNodeInMapping(currentNode, varName)
	if keyNode == nil {
		return nil, fmt.Errorf("variable not found in section")
	}

	// Return 0-indexed position
	return &Location{
		File: configPath,
		Line: keyNode.Line - 1,
		Col:  keyNode.Column - 1,
	}, nil
}

// FindNodeInMapping finds the value node for a given key in a YAML mapping.
func FindNodeInMapping(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	// MappingNode content is [key, value, key, value, ...]
	for i := 0; i < len(node.Content); i += 2 {
		if i+1 < len(node.Content) && node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}

	return nil
}

// FindKeyNodeInMapping finds the key node itself (not the value) in a YAML mapping.
func FindKeyNodeInMapping(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	// MappingNode content is [key, value, key, value, ...]
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i]
		}
	}

	return nil
}

// FindSectionPosition finds the position of a section path in a YAML file.
// Returns the position of the last key in the path.
func FindSectionPosition(projectRoot string, section []string) (*Location, error) {
	if len(section) == 0 {
		return nil, fmt.Errorf("empty section path")
	}

	// The last element is the key we want to find
	varName := section[len(section)-1]
	parentSection := section[:len(section)-1]

	return FindVarPositionInYAML(projectRoot, varName, parentSection)
}
