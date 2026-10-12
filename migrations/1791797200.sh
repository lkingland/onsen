echo "Hide the bar until Super+Shift+Space shows it"

# One shot for homes that already exist. After this, the toggle sticks.
# Same path the shell watches: ~/.local/state, not $XDG_STATE_HOME.
flag="$HOME/.local/state/omarchy/toggles/bar-off"
mkdir -p "$(dirname "$flag")"
if [[ ! -e "$flag" ]]; then
  : > "$flag"
fi
