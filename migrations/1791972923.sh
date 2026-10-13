echo "Source the Onsen bash config from ~/.bashrc"

# The shipped .bashrc (default/bashrc) carries this line; a home that
# already has its own gets it once.
if ! grep -q 'config/bash/rc' ~/.bashrc 2>/dev/null; then
  printf '\n# Onsen edition: its aliases, functions, and tool wrappers\n[[ -r ~/.config/bash/rc ]] && source ~/.config/bash/rc\n' >> ~/.bashrc
fi
