echo "Onsen bar: bottom edge, overlay, opaque background"

# The edition ships config/omarchy/shell.json, but onsen update seeds it only
# when none exists (default/onsen/drift-ignore), and Omarchy's first login
# writes one first. So a fresh machine starts with the stock bar on the top
# edge. This applies the edition's bar settings once, the way omarchy bar does.
omarchy-bar position bottom
omarchy-bar background opaque
source omarchy-shell-config
commit "$NORMALIZE | .bar.overlay = true"
