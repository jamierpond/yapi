# Source this file to enable yapi shell history integration.
# After running yapi, the command will be added to your zsh history.

function yapi() {
  command yapi "$@"
  local success=$?
  [ $success -ne 0 ] && return $success

  local file="$HOME/.yapi_history.json"
  [ -f "$file" ] || return 0

  local recent_line=$(tail -n 1 "$file")
  local cmd=$(echo "$recent_line" | jq -r '.command' 2>/dev/null)
  [ -n "$cmd" ] && { print -s "$cmd" 2>/dev/null || history -s "$cmd" 2>/dev/null; }
}


