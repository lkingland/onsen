-- Keep only your personal keybinding overrides here. Add new bindings with
-- o.bind or replace defaults with o.rebind.

-- See current bindings and descriptions:
--   omarchy menu keybindings --print

-- To disable every Omarchy default binding, set this in
-- ~/.config/hypr/hyprland.lua before require("default.hypr.omarchy"), then add
-- only the bindings you want below:
--   omarchy_default_bindings = false

-- To disable all preinstalled app/webapp bindings, set:
--   omarchy_preinstalled_bindings = false

-- Add a new binding.
-- o.bind("SUPER + SHIFT + R", "SSH", "alacritty -e ssh your-server")

-- Change an existing binding. o.rebind takes the same arguments as o.bind.
-- This example replaces the default file manager with Flea.
-- o.rebind("SUPER + SHIFT + F", "File manager", { launch = "flea" })

-- Disable a default binding without replacing it.
-- hl.unbind("SUPER + SHIFT + B")

-- Logitech MX Keys examples:
-- o.bind("SUPER + H", nil, "voxtype record toggle")
-- o.bind("SUPER + PERIOD", nil, { panel = "omarchy.emojis" })

-- Onsen: vim keys. H J K L focus a window, Shift + H J K L move it.
-- J was the split toggle, now Ctrl + J. K was the keybindings menu, now
-- Shift + Ctrl + K (Ctrl + K stays Herdr). L was the workspace-layout
-- cycle, which loses its key, and Alt + L with it. Q, a second close, is
-- the previous workspace and E the next on this monitor, as Super + Tab is
-- upstream; W still closes. Super + number is upstream's too: each monitor
-- has workspaces 1-10 of its own (default/hypr/monitor-workspaces.lua).
o.rebind("SUPER + Q", "Previous workspace", hl.dsp.focus({ workspace = "m-1" }))
o.bind("SUPER + E", "Next workspace", hl.dsp.focus({ workspace = "m+1" }))
o.rebind("SUPER + J", "Focus on below window", hl.dsp.focus({ direction = "d" }))
o.bind("SUPER + CTRL + J", "Toggle window split", hl.dsp.layout("togglesplit"))
o.rebind("SUPER + K", "Focus on above window", hl.dsp.focus({ direction = "u" }))
o.bind("SUPER + SHIFT + CTRL + K", "Keybindings", "omarchy-menu-keybindings")
o.rebind("SUPER + L", "Focus on right window", hl.dsp.focus({ direction = "r" }))
hl.unbind("SUPER + ALT + L")
o.bind("SUPER + H", "Focus on left window", hl.dsp.focus({ direction = "l" }))
o.bind("SUPER + SHIFT + H", "Move window left", hl.dsp.window.move({ direction = "l" }))
o.bind("SUPER + SHIFT + J", "Move window down", hl.dsp.window.move({ direction = "d" }))
o.bind("SUPER + SHIFT + K", "Move window up", hl.dsp.window.move({ direction = "u" }))
o.bind("SUPER + SHIFT + L", "Move window right", hl.dsp.window.move({ direction = "r" }))

-- Super + Ctrl + number focuses monitor N (Hyprland's ids, in enumeration
-- order, from 1), the pair to Super + number being a workspace on a monitor.
-- Stock has these open the bar's panels by position; the letter binds remain.
for index = 1, 9 do
  o.rebind("SUPER + CTRL + code:" .. tostring(index + 9), "Focus monitor " .. index, hl.dsp.focus({ monitor = tostring(index - 1) }))
end

-- Hold Super to see where focus is: the onsen.beacon shell plugin
-- (config/omarchy/plugins/onsen.beacon) pulses at the focused window while
-- Super is down, and shows only after a short hold, so chords never flash
-- it. bind_hold's release half fires even when other keys were pressed
-- during the hold. Non-consuming, so Super still reaches everything else.
o.bind_hold(
  "SUPER + Super_L",
  "Show the focus beacon", "omarchy-shell shell summon onsen.beacon '{}'",
  "Hide the focus beacon", "omarchy-shell shell hide onsen.beacon",
  { non_consuming = true }
)
