echo "Set up Brave the way omarchy install browser brave does, and make it the default"

# The installer brings the Wayland flags, policies, Widevine, and the theme
# hook, then leaves the default browser alone (Setup > Defaults > Browser).
# The edition's default is Brave, for its shields, so it is set here too: the
# web apps (X, YouTube, SoundCloud) open in whatever the default is.
omarchy-install-browser brave
omarchy-default-browser brave
