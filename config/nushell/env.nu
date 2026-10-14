# env.nu
#
# Installed by:
# version = "0.108.0"
#
# Previously, environment variables were typically configured in `env.nu`.
# In general, most configuration can and should be performed in `config.nu`
# or one of the autoload directories.
#
# This file is generated for backwards compatibility for now.
# It is loaded before config.nu and login.nu
#
# See https://www.nushell.sh/book/configuration.html
#
# Also see `help config env` for more options.
#
# You can remove these comments if you want or leave
# them for future reference.

# Locale settings for proper UTF-8 support (required for tmux, mosh, etc.)
$env.LANG = "en_US.UTF-8"
$env.LC_CTYPE = "en_US.UTF-8"
$env.LC_ALL = "en_US.UTF-8"

# XDG Base Directory specification
$env.XDG_CONFIG_HOME = $"($env.HOME)/.config"
$env.XDG_DATA_HOME = $"($env.HOME)/.local/share"
$env.XDG_CACHE_HOME = $"($env.HOME)/.cache"

# Omarchy paths (required by omarchy-version and other scripts)
$env.OMARCHY_PATH = "/usr/share/omarchy"

# Pager settings (better git diff, man pages, etc.)
$env.LESS = "-R --no-init --quit-if-one-screen"

# Kubernetes editor
$env.KUBE_EDITOR = "nvim"

# Carapace completion - bridges completions from multiple shells
$env.CARAPACE_BRIDGES = 'zsh,fish,bash,inshellisense'
mkdir ($nu.cache-dir)
carapace _carapace nushell | save --force $"($nu.cache-dir)/carapace.nu"
