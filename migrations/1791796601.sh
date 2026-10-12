echo "Build the bat cache for the Onsen theme"

theme="${XDG_CONFIG_HOME:-$HOME/.config}/bat/themes/onsen.tmTheme"
if command -v bat >/dev/null 2>&1 && [[ -f "$theme" ]]; then
  bat cache --build
fi
