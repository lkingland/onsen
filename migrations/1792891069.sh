echo "Retire the hand-made Bluetooth dongle files the Onsen package now carries"

# etc/modprobe.d/bluetooth.conf (from omarchy-settings) now turns off btusb's
# USB autosuspend, which two hand-made files on aurora did until now. Remove
# them only once the packaged file carries the option, so the dongle is never
# left able to suspend in between.
if grep -qs '^options btusb enable_autosuspend=0' /etc/modprobe.d/bluetooth.conf; then
  sudo rm -f /etc/modprobe.d/btusb-no-autosuspend.conf /etc/udev/rules.d/50-bluetooth-dongle-no-autosuspend.rules
fi

# The retired connectable rule and the new link-policy rule arrive with the
# package; make sure udev reads the directory as it stands now.
sudo udevadm control --reload

# A trackpad already connected got its link before the rule existed. Set its
# policy now rather than waiting for the next reconnect; nothing to do when no
# input device is linked.
if systemctl is-active --quiet bluetooth.service; then
  sudo omarchy-bluetooth-link-policy >/dev/null 2>&1 || true
fi
