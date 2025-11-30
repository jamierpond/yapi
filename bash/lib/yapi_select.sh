#!/bin/bash
# yapi_select.sh - FZF-based config file selection

# Select config file using fzf
select_config_file() {
  local use_all_files="$1"
  local selected_config=""

  check_dependency "fzf" "Config file is required (use -c or --config). Install fzf for interactive selection."

  local yaml_files=""

  if [[ "$use_all_files" == "true" ]]; then
    # Search all YAML files in directory tree
    yaml_files=$(find . -type f \( -name '*.yml' -o -name '*.yaml' \) 2>/dev/null | sed 's|^\./||')
    if [[ -z "$yaml_files" ]]; then
      error_exit "No YAML files found in directory tree"
    fi
    selected_config=$(echo "$yaml_files" | fzf --preview 'yq {}' --prompt="Select config file: ")
  else
    # Search only git-tracked YAML files (default)
    yaml_files=$(git ls-files "*.$YAPI_EXTENSION.yaml" "*.$YAPI_EXTENSION.yml" 2>/dev/null)
    if [[ -z "$yaml_files" ]]; then
      error_exit "No git-tracked *.$YAPI_EXTENSION.[yaml|yml] files found. Use --all to search all files in directory tree."
    fi
    selected_config=$(echo "$yaml_files" | fzf --preview 'yq {}' --prompt="Select config file: ")
  fi

  if [[ -z "$selected_config" ]]; then
    error_exit "No config file selected"
  fi

  echo "Selected: $selected_config" >&2
  echo "$selected_config"
}
