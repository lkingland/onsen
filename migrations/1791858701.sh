echo "Retire the hand-made sudo grant the Onsen package now carries"

# /etc/sudoers.d/50-onsen-agent-sudo (from omarchy-settings) grants what the
# hand-made ssh-nopasswd did. Remove the old file only once the new one is in
# place, so sudo never loses the grant in between.
if sudo test -f /etc/sudoers.d/50-onsen-agent-sudo && sudo test -e /etc/sudoers.d/ssh-nopasswd; then
  sudo rm -f /etc/sudoers.d/ssh-nopasswd
fi
