echo "Install GoMono Nerd Font"

if ! pacman -Qq ttf-go-nerd >/dev/null 2>&1; then
  omarchy-pkg-add ttf-go-nerd
fi
