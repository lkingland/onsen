# Build bat's theme cache once the Onsen tmTheme is in the config dir.
# No-op when bat or the theme file is absent.
set -euo pipefail

theme="${XDG_CONFIG_HOME:-$HOME/.config}/bat/themes/onsen.tmTheme"
if command -v bat >/dev/null 2>&1 && [[ -f "$theme" ]]; then
  bat cache --build
fi
