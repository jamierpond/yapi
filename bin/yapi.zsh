# Source this file to enable yapi shell history integration.
# After running yapi, the command will be added to your zsh history.
# This is especially useful for TUI mode - select a file interactively,
# then press up arrow to re-run the equivalent CLI command.

function yapi() {
  local file="$HOME/.yapi/history.json"
  local before_count=0
  [ -f "$file" ] && before_count=$(wc -l < "$file" | tr -d ' ')

  command yapi "$@"
  local exit_code=$?

  # Check if a new entry was added
  [ -f "$file" ] || return $exit_code
  local after_count=$(wc -l < "$file" | tr -d ' ')
  [ "$after_count" -gt "$before_count" ] || return $exit_code

  # Add the new command to shell history
  local cmd=$(tail -n 1 "$file" | jq -r '.command' 2>/dev/null)
  [ -n "$cmd" ] && { print -s "$cmd" 2>/dev/null || history -s "$cmd" 2>/dev/null; }

  return $exit_code
}


