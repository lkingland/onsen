# config.nu - Nushell configuration
#
# Nushell quick reference:
# - `help commands` - list all commands
# - `help <command>` - get help for a specific command
# - `$env` - view environment variables
# - `config nu` - edit this file
# - Pipelines work like unix but pass structured data, not text

# -------------------------
# Settings
# -------------------------
$env.config = {
    show_banner: false
    edit_mode: vi

    cursor_shape: {
        vi_insert: line
        vi_normal: block
    }

    history: {
        max_size: 10000000
        sync_on_enter: true
        file_format: "sqlite"
        isolation: true
    }

    completions: {
        case_sensitive: false
        quick: true
        partial: true
        algorithm: "fuzzy"
    }

    # Onsen color config - muted colors for data, colors only for actionable items

    keybindings: [
        {
            name: clear_screen
            modifier: control
            keycode: char_l
            mode: [emacs, vi_normal, vi_insert]
            event: { send: clearscreen }
        }
        {
            name: ide_completion
            modifier: control
            keycode: char_n
            mode: [emacs, vi_normal, vi_insert]
            event: { send: menu name: ide_completion_menu }
        }
    ]
}

# Vi mode indicators (starship character disabled, nushell handles it)

# -------------------------
# Environment
# -------------------------

# Add local bin directories to PATH.
# Omarchy's binaries are /usr/bin/omarchy-* from the package.
$env.PATH = ($env.PATH | prepend $"($env.HOME)/.local/bin")
$env.PATH = ($env.PATH | prepend $"($env.HOME)/go/bin")
$env.PATH = ($env.PATH | prepend $"($env.HOME)/sh")

# Java (use `sudo archlinux-java set java-21-openjdk` for Android)
$env.JAVA_HOME = "/usr/lib/jvm/default"

# Android SDK
$env.ANDROID_HOME = $"($env.HOME)/Android/Sdk"
$env.ANDROID_SDK_ROOT = $env.ANDROID_HOME
$env.PATH = ($env.PATH | append $"($env.ANDROID_HOME)/platform-tools")
$env.PATH = ($env.PATH | append $"($env.ANDROID_HOME)/cmdline-tools/latest/bin")
$env.PATH = ($env.PATH | append $"($env.ANDROID_HOME)/emulator")
$env.PATH = ($env.PATH | append $"($env.ANDROID_HOME)/build-tools/34.0.0")

$env.EDITOR = "nvim"
# Quattro does not ship default/bash/tools. These bw/jira/linode wrappers live
# with the other dotfiles so a package update cannot take them.
$env.BASH_ENV = $"($env.HOME)/.config/bash/tools"

# Blue=structure, White=text, Green=binary, Orange=executable

# GPG agent for SSH authentication (YubiKey)
# smartsocket owns S.gpg-agent{,.ssh} and routes to either the
# SSH-forwarded remote agent or a socket-activated local gpg-agent.
# Do NOT launch a free-running gpg-agent here — it would steal the
# socket files away from smartsocket on rebind.
$env.GPG_TTY = (tty)
$env.SSH_AUTH_SOCK = "/run/user/1000/gnupg/S.gpg-agent.ssh"
$env.VISUAL = "nvim"
$env.SUDO_EDITOR = "nvim"
$env.GIT_EXTERNAL_DIFF = "difft"
$env.DFT_UNDERLINE = "false"
$env.TIME_STYLE = "+%Y/%m/%d %H:%M"

# -------------------------
# Aliases - General
# -------------------------
alias fg = job unfreeze
alias x = clear
alias e = exit
alias v = nvim
alias s = sudo
alias d = docker
alias g = git
alias h = hamr
alias hd = hamr-dissolve
alias hp = hamr-projects
alias hs = hamr-spawn
alias ht = hamr-status
alias k = kubectl

# Create directory and cd into it
def cx [dir: string] {
    mkdir $dir
    cd $dir
}

# Nvim with directory default (like Omarchy's n function)
def n [...args: string] {
    if ($args | is-empty) {
        nvim .
    } else {
        nvim ...$args
    }
}

# Open file/URL in background (xdg-open wrapper)
def open [path: string] {
    xdg-open $path out+err> /dev/null &
}

# Directory stack (pushd/popd style navigation)
$env.DIR_STACK = []

# cd wrapper that pushes current dir onto stack before changing
def --env c [dir?: string] {
    let target = if ($dir | is-empty) { $env.HOME } else { $dir }
    $env.DIR_STACK = ($env.DIR_STACK | prepend $env.PWD)
    cd $target
}

# Pop directory from stack and cd to it
def --env o [] {
    if ($env.DIR_STACK | is-empty) {
        print "Directory stack is empty"
    } else {
        let target = ($env.DIR_STACK | first)
        $env.DIR_STACK = ($env.DIR_STACK | skip 1)
        cd $target
    }
}

# Show the directory stack
def dirs [] {
    $env.DIR_STACK | enumerate | each { |it| $"($it.index): ($it.item)" }
}

# Directory navigation shortcuts
alias '..' = cd ..
alias '...' = cd ../..
alias '....' = cd ../../..

# -------------------------
# Aliases - ls variants
# -------------------------
# Eza (compact, readable output)
alias l = eza -l --icons --git
alias ll = eza -la --icons --git
alias lla = eza -la --icons --git
alias la = eza -a --icons

# Eza tree variants
alias lt = eza --tree --level=2 --long --icons --git
alias lta = eza --tree --level=2 --long --icons --git -a
alias ltree = eza --tree --level=2 --icons --git
alias ltt = eza --tree --level=4 --icons --git --only-dirs

# Nushell native ls (for structured data / piping with | where, etc.)
# Use `ls` directly for native output
alias lsa = ls -a

# Fuzzy finder with bat preview
def ff [...args: string] {
    if ($args | is-empty) {
        fzf --preview 'bat --color=always {}'
    } else {
        fzf --preview 'bat --color=always {}' ...$args
    }
}

# -------------------------
# Aliases - Git
# -------------------------
alias ga = git add
alias gb = git branch
alias gba = do { git fetch -ap; git branch -a }
alias gc = git commit
alias gca = git commit -a
alias gco = git checkout
alias gd = git diff
alias gp = git pull
alias gr = git remote -v
alias gs = git status -sb
alias gt = git tag --sort=version:refname
alias gtd = git tag --delete
alias gu = git push

# Git commit shortcuts (from Omarchy)
def gcm [message: string] { git commit -m $message }
def gcam [message: string] { git commit -a -m $message }
def gcad [] { git commit -a --amend }

# Git log - compact format
def gl [n: int = 10] {
    git log -n $n --graph --pretty=format:"%Cgreen%h%Creset %an: %s %Creset %C(yellow)%d%Creset %Cgreen(%cr)%Creset" --abbrev-commit --date=relative
}

# Git log - all commits
def gll [] {
    git log --graph --pretty=format:"%Cgreen%h%Creset %an: %s %Creset %C(yellow)%d%Creset %Cgreen(%cr)%Creset" --abbrev-commit --date=relative
}

# Git log - fancy topo-ordered (from fish)
def glog [] {
    git log --graph --topo-order --pretty=format:"%w(100,0,6)%C(yellow)%h%C(bold)%C(black)%d %C(cyan)%ar %C(green)%an%n%C(bold)%C(white)%s %N" --abbrev-commit
}

# Difftastic aliases (fancy structural diffs)
def dl [...args: string] { git log -p --ext-diff ...$args }
def ds [...args: string] { git show --ext-diff ...$args }
def dt [...args: string] { git diff --ext-diff ...$args }

# -------------------------
# Aliases - Systemd
# -------------------------
alias sc = systemctl
alias scu = systemctl --user
alias jc = journalctl --output=cat
alias jcu = journalctl --output=cat --user

# -------------------------
# Aliases - Wayland clipboard
# -------------------------
alias copy = wl-copy
alias paste = wl-paste

# -------------------------
# Jujutsu VCS
# -------------------------
def jll [] {
    jj log -r '@ | ancestors(remote_bookmarks().., 2) | trunk()'
}

# -------------------------
# Bitwarden (session management)
# -------------------------
def bw-unlock [] {
    ^bw unlock --raw | save -f ~/.config/bw/session
}

def --wrapped bw [...args] {
    let session = (^cat ~/.config/bw/session | str trim)
    ^bw --session $session ...$args
}

# Jira CLI (token from Bitwarden)
def --wrapped jira [...args] {
    with-env { JIRA_API_TOKEN: (bw get password "jira token" | str trim) } {
        ^jira ...$args
    }
}

# -------------------------
# Utility Functions
# -------------------------

# Compression utilities (from Omarchy)
def compress [dir: string] {
    let name = ($dir | path basename)
    tar -czvf $"($name).tar.gz" $dir
}

def decompress [file: string] {
    tar -xzvf $file
}

# Media transcoding (from Omarchy)
def transcode-video-1080p [input: string, output: string] {
    ffmpeg -i $input -vf "scale=1920:1080" -c:v libx264 -crf 23 -c:a aac -b:a 128k $output
}

def transcode-video-4K [input: string, output: string] {
    ffmpeg -i $input -vf "scale=3840:2160" -c:v libx265 -crf 28 -c:a aac -b:a 192k $output
}

def img2jpg [input: string, output: string] {
    magick $input -quality 85 $output
}

def img2jpg-small [input: string, output: string] {
    magick $input -resize "1920x1080>" -quality 75 $output
}

def img2png [input: string, output: string] {
    magick $input -compress lossless $output
}

# ISO to SD card utility
def iso2sd [iso: string, device: string] {
    print $"WARNING: This will overwrite ($device)!"
    print "Press Ctrl+C to cancel, or wait 5 seconds to continue..."
    sleep 5sec
    sudo dd if=$iso of=$device bs=4M status=progress oflag=sync
}

# -------------------------
# Integrations
# -------------------------

# Mise (version manager) - CRITICAL for managing tool versions
# Regenerate with: mise activate nu | save -f ~/.cache/mise.nu
source ~/.cache/mise.nu

# Zoxide (smart cd) - use 'z' to jump to directories
source ~/.cache/zoxide.nu

# Override z to push current dir onto stack before jumping
def --env --wrapped z [...rest: string] {
    $env.DIR_STACK = ($env.DIR_STACK | prepend $env.PWD)
    __zoxide_z ...$rest
}

# Starship prompt
source ~/.cache/starship.nu

# Carapace completions (rich completions for 1000+ commands)
source $"($nu.cache-dir)/carapace.nu"

# Atuin (fish-style history with fuzzy search)
# Ctrl+R for fuzzy search, deduplication, shared history
source ~/.cache/atuin.nu

# -------------------------
# Secrets (Bitwarden CLI)
# -------------------------

# Linode CLI — agent/ops high-auth PAT from bw (no Object Storage write).
# Backup OS mutation is linode-backups-token only; never wire it here.
# Token lives in env only for the call duration (not ~/.config/linode-cli).
# Requires `bw unlock` first. See wiki Present/I12E/Linode-Tokens.md.
def --wrapped linode-cli [...args] {
    with-env { LINODE_CLI_TOKEN: (bw get password linode-cli-token) } {
        ^linode-cli ...$args
    }
}

# Power-verb guardrail (2026-07-31, operator-approved via Hamr#412 arc):
# a dead ssh session silently drops power commands into the LOCAL shell —
# the wrong-pane class. Make the blast host explicit: name it to proceed.
# Interactive shells only by construction (config.nu); agents' non-interactive
# bash is untouched — this is a failsafe, not a gate. Deliberate bypass when
# you really mean it: ^poweroff (caret calls the external directly).
def power-verb-guard [verb: string, ...args] {
  let host = ((sys host).hostname | str uppercase)
  print $"(ansi red_bold)⚠ You are on ($host)(ansi reset) — ($verb) will act HERE."
  let resp = (try { input $"Type '($host)' to proceed: " } catch { "" })
  if ($resp | str uppercase) == $host {
    run-external $verb ...$args
  } else {
    print $"(ansi yellow)($verb) aborted — hostname not confirmed.(ansi reset)"
  }
}
def --wrapped poweroff [...args] { power-verb-guard poweroff ...$args }
def --wrapped reboot   [...args] { power-verb-guard reboot ...$args }
def --wrapped shutdown [...args] { power-verb-guard shutdown ...$args }
