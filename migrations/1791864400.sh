echo "Let Sunshine's streaming ports in, on the tailnet only"

# Sunshine streams this desktop to Moonlight. Its package is declared in
# default/onsen/packages and its service in default/onsen/services. Like
# SSH, it is reachable over tailscale0 only. The web UI on 47990 asks for
# the credentials kept in ~/state/sunshine.
if omarchy-cmd-present ufw; then
  for port in 47984 47989 47990 48010; do
    sudo ufw allow in on tailscale0 to any port "$port" proto tcp comment "onsen-sunshine-tailnet" >/dev/null
  done
  for port in 47998:48000 48002 48010; do
    sudo ufw allow in on tailscale0 to any port "$port" proto udp comment "onsen-sunshine-tailnet" >/dev/null
  done
  sudo ufw reload >/dev/null
fi
