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
-- the previous workspace and E the next; W still closes.
o.rebind("SUPER + Q", "Previous workspace", hl.dsp.focus({ workspace = "e-1" }))
o.bind("SUPER + E", "Next workspace", hl.dsp.focus({ workspace = "e+1" }))
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

-- Super + number is the Nth workspace of the focused monitor, not a global id:
-- Hyprland's r~N selector counts the current monitor's workspaces, empty or
-- not yet created included. Ids stay global underneath, so the bar's numbers
-- are the ids, not these positions.
for workspace = 1, 10 do
  local key = "code:" .. tostring(workspace + 9)
  local target = "r~" .. tostring(workspace)
  o.rebind("SUPER + " .. key, "Switch to workspace " .. workspace .. " on this monitor", hl.dsp.focus({ workspace = target }))
  o.rebind("SUPER + SHIFT + " .. key, "Move window to workspace " .. workspace .. " on this monitor", hl.dsp.window.move({ workspace = target }))
  o.rebind("SUPER + SHIFT + ALT + " .. key, "Move window silently to workspace " .. workspace .. " on this monitor", hl.dsp.window.move({ workspace = target, follow = false }))
end

-- Super + Ctrl + number focuses monitor N (Hyprland's ids, in enumeration
-- order, from 1), the pair to Super + number being a workspace on a monitor.
-- Stock has these open the bar's panels by position; the letter binds remain.
for index = 1, 9 do
  o.rebind("SUPER + CTRL + code:" .. tostring(index + 9), "Focus monitor " .. index, hl.dsp.focus({ monitor = tostring(index - 1) }))
end
