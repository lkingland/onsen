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

-- Hold Super to see where focus is. After a short hold, so Super chords
-- never flash it, the focused window's border shows in green even when it
-- is alone on its screen (looknfeel.lua keeps that border, transparent),
-- and the onsen.beacon shell plugin (config/omarchy/plugins/onsen.beacon)
-- pulses at its centre. Releasing Super puts both away.
--
-- The key is plain Super_L, with no modifier in the bind: Hyprland matches
-- SUPER + Super_L only once Super already counts as held, which is at the
-- release, so that bind's press half fired together with its release
-- (measured 2026-10-19: both within 5 ms of the release). bind_hold's
-- release half ignores modifiers and fires even when other keys were
-- pressed during the hold. Non-consuming, so Super still reaches everything.
local onsen_focus = { held = false }

local function onsen_focus_press()
  onsen_focus.held = true
  hl.exec_cmd("omarchy-shell shell summon onsen.beacon '{}'")
  hl.timer(function()
    if onsen_focus.held then onsen_lone_borders(true) end
  end, { timeout = 220, type = "oneshot" })
end

local function onsen_focus_release()
  onsen_focus.held = false
  onsen_lone_borders(false)
  hl.exec_cmd("omarchy-shell shell hide onsen.beacon")
end

o.bind_hold(
  "Super_L",
  "Show where focus is", onsen_focus_press,
  "Put the focus marks away", onsen_focus_release,
  { non_consuming = true }
)
