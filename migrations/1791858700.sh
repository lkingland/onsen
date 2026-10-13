echo "Turn off systemd-oomd, which would treat the agent fleet as one app"

# Stock Omarchy enables systemd-oomd (enable-services.sh, migration 1785424256)
# and lets it kill app.slice under memory pressure. On an Onsen machine the
# tmux server and every agent pane run in app.slice, so one kill there ends
# the whole fleet. Masked, not just disabled, so a later stock step that
# enables it again does nothing.
if ! sudo systemctl mask --now systemd-oomd.service systemd-oomd.socket >/dev/null 2>&1; then
  echo "Could not mask systemd-oomd; memory pressure can still end the fleet."
fi
