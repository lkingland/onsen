echo "Allow SSH and mosh in from the tailnet only"

# The Onsen sshd settings ship in /etc/ssh/sshd_config.d/50-onsen.conf. Stock
# Omarchy turns the server on only through omarchy-setup-security-sshd, which
# opens port 22 on every interface. Here it is reachable over tailscale0 only.
if omarchy-cmd-present sshd; then
  sudo systemctl enable --now sshd.service
fi

if omarchy-cmd-present ufw; then
  sudo ufw allow in on tailscale0 to any port 22 proto tcp comment "onsen-ssh-tailnet" >/dev/null
  sudo ufw allow in on tailscale0 to any port 60000:61000 proto udp comment "onsen-mosh-tailnet" >/dev/null
  sudo ufw reload >/dev/null
fi
