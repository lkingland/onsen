echo "Install tmux's plugin manager and the plugins tmux.conf names"

# tmux.conf runs TPM and lists its plugins; nothing else puts them in place,
# and without them a reload fails and vim-tmux-navigator never takes C-h/j/k/l.
# This runs inside the stock update, before the edition's tmux.conf reaches
# ~/.config, so TPM's own installer would read stock's config and find no
# plugins. The list is read from the edition's copy of the file instead, and
# each plugin cloned where TPM keeps them. The plugins directory is onsen's
# drift-ignore; TPM owns it from here.
conf="$OMARCHY_PATH/config/tmux/tmux.conf"
plugins="$HOME/.config/tmux/plugins"
mkdir -p "$plugins"
while read -r repo; do
  [[ -n $repo ]] || continue
  name=${repo##*/}
  if [[ ! -d $plugins/$name ]]; then
    git clone -q "https://github.com/$repo" "$plugins/$name"
  fi
done < <(sed -nE "s/^set -g @plugin '([^']+)'.*/\1/p" "$conf")
