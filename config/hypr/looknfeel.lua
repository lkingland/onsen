-- Change the default Omarchy look'n'feel.

-- Opt another application in to Omarchy's standard transparency.
-- Find its class with: hyprctl clients
-- o.transparent_window("my-app")
-- o.transparent_window("my-app", "0.9 0.85") -- Custom active/inactive opacity.

-- https://wiki.hypr.land/Configuring/Basics/Variables/#general
-- hl.config({
--   general = {
--     -- No gaps between windows or borders.
--     gaps_in = 0,
--     gaps_out = 0,
--     border_size = 0,
--
--     -- Change to niri-like side-scrolling layout.
--     layout = "scrolling",
--   },
-- })

-- https://wiki.hypr.land/Configuring/Basics/Variables/#decoration
-- hl.config({
--   decoration = {
--     -- Use round window corners.
--     rounding = 8,
--
--     -- Dim unfocused windows (0.0 = no dim, 1.0 = fully dimmed).
--     dim_inactive = true,
--     dim_strength = 0.15,
--   },
-- })

-- https://wiki.hypr.land/Configuring/Basics/Variables/#animations
-- hl.config({
--   animations = {
--     -- Disable all animations.
--     enabled = false,
--   },
-- })

-- https://wiki.hypr.land/Configuring/Basics/Variables/#layout
-- hl.config({
--   layout = {
--     -- Avoid overly wide single-window layouts on wide screens.
--     single_window_aspect_ratio = { 1, 1 },
--   },
-- })

-- https://wiki.hypr.land/Configuring/Layouts/Scrolling-Layout/
-- hl.config({
--   scrolling = {
--     -- See only one column per screen instead of two.
--     column_width = 0.97,
--   },
-- })

-- A small outer gap. These desks are LG TVs, whose edges do not land
-- exactly on the panel's, so a window flush with the edge can be clipped;
-- 5 keeps it clear. The inner gap stays stock.
hl.config({ general = { gaps_out = 5 } })

-- A lone tiled window, or a workspace that is one group, shows no border:
-- on a 48" screen a frame around the only window is a garish outline with
-- nothing to tell apart. The border is still there, at the stock size, only
-- transparent, so holding Super can show it (bindings.lua) without moving
-- the window. Borders show as soon as a second window arrives.
function _G.onsen_lone_borders(visible)
  hl.window_rule({
    name = "onsen-lone-border",
    enabled = not visible,
    match = { float = false, workspace = "w[tv1]" },
    border_color = "rgba(00000000)",
  })
  hl.window_rule({
    name = "onsen-lone-group-border",
    enabled = not visible,
    match = { float = false, workspace = "w[tg1]" },
    border_color = "rgba(00000000)",
  })
end
onsen_lone_borders(false)

-- Bar reveal. The bar is kept hidden (the bar-off toggle; Super + Shift +
-- Space shows it) and sits on the bottom edge. Resting the pointer on that
-- edge brings it back for as long as the pointer stays on the bar's strip;
-- moving away hides it again. Only a bar this revealed is hidden by it: one
-- shown by hand stays. A compositor timer polls the pointer in-process, ten
-- times a second, since Hyprland raises no event for pointer motion and a
-- layer-shell strip under the hidden bar never saw hover here.
local onsen_bar_reveal = { shown_by_us = false }

local function onsen_bar_hidden()
  local f = io.open((os.getenv("HOME") or "") .. "/.local/state/omarchy/toggles/bar-off", "r")
  if f then
    f:close()
    return true
  end
  return false
end

local function onsen_bar_set_hidden(hidden)
  hl.exec_cmd("omarchy-toggle bar-off " .. (hidden and "on" or "off") .. " && omarchy-shell -q omarchy.bar syncHidden")
end

local function onsen_bar_tick()
  local c = hl.get_cursor_pos()
  local m = hl.get_monitor_at_cursor()
  if c and m and m.scale and m.scale > 0 then
    local bottom = m.y + m.height / m.scale
    local bar = 26 -- the bar's horizontal size, as Style.bar.sizeHorizontal
    if not onsen_bar_reveal.shown_by_us then
      if c.y >= bottom - 2 and onsen_bar_hidden() then
        onsen_bar_reveal.shown_by_us = true
        onsen_bar_set_hidden(false)
      end
    elseif onsen_bar_hidden() then
      -- Hidden by hand meanwhile: not ours any more.
      onsen_bar_reveal.shown_by_us = false
    elseif c.y < bottom - bar - 8 then
      onsen_bar_reveal.shown_by_us = false
      onsen_bar_set_hidden(true)
    end
  end
  hl.timer(onsen_bar_tick, { timeout = 100, type = "oneshot" })
end

if not _G.onsen_bar_reveal_started then
  _G.onsen_bar_reveal_started = true
  hl.timer(onsen_bar_tick, { timeout = 1000, type = "oneshot" })
end

-- Onsen: the focused window's frame in the Onsen palette's green, the same as
-- tmux's active pane border, in place of the theme's accent. This file loads
-- after the theme's colours, so it holds across a theme switch.
hl.config({
  general = { col = { active_border = "#8FD9A8" } },
  group = { col = { border_active = "#8FD9A8" } },
})
