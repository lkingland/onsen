-- Keep only your personal input overrides here. Uncommented settings below
-- replace Omarchy's defaults.

-- Keyboard layout and options.
-- See https://wiki.hypr.land/Configuring/Basics/Variables/#input
-- hl.config({
--   input = {
--     -- Use multiple keyboard layouts and switch between them with Left Alt + Right Alt.
--     kb_layout = "us,dk,eu",
--     kb_options = "omarchy:shift_caps_compose,grp:alts_toggle",
--
--     -- Use a specific keyboard variant if needed (e.g. intl for international keyboards).
--     kb_variant = "intl",
--
--     -- Change speed of keyboard repeat.
--     repeat_rate = 40,
--     repeat_delay = 250,
--
--     -- Start with numlock on by default.
--     numlock_by_default = true,
--
--     -- Increase sensitivity for mouse/trackpad (default: 0).
--     sensitivity = 0.35,
--
--     -- Turn off mouse acceleration (default: adaptive).
--     accel_profile = "flat",
--
--     touchpad = {
--       -- Use natural (inverse) scrolling.
--       natural_scroll = true,
--
--       -- Use two-finger clicks for right-click instead of lower-right corner.
--       clickfinger_behavior = true,
--
--       -- Control the speed of your scrolling.
--       scroll_factor = 0.4,
--
--       -- Enable the touchpad while typing.
--       disable_while_typing = false,
--
--       -- Left-click-and-drag with three fingers.
--       drag_3fg = 1,
--     },
--   },
-- })

-- App-specific touchpad scroll speeds.
-- o.window("(Alacritty|kitty)", { scroll_touchpad = 1.5 })
-- o.window("foot", { scroll_touchpad = 2.0 })
-- o.window("com.mitchellh.ghostty", { scroll_touchpad = 0.2 })

-- Enable touchpad gestures for changing workspaces.
-- See https://wiki.hypr.land/Configuring/Advanced-and-Cool/Gestures/
-- hl.gesture({ fingers = 3, direction = "horizontal", action = "workspace" })

-- Enable touchpad gestures for moving focus (helpful on scrolling layout).
-- hl.gesture({ fingers = 3, direction = "left", action = function() hl.dispatch(hl.dsp.focus({ direction = "l" })) end })
-- hl.gesture({ fingers = 3, direction = "right", action = function() hl.dispatch(hl.dsp.focus({ direction = "r" })) end })

-- Onsen: compose on Right Alt (stock: Shift + Caps Lock), so a layout switch
-- is Alt + Shift; a quicker key repeat; Apple-style natural scrolling, drag
-- lock (libinput's timeout mode), and tap-and-drag on the trackpad; and
-- three-finger workspace swipes that cross several workspaces and open a
-- new one past the last.
hl.config({
  input = {
    kb_options = "compose:ralt,grp:alt_shift_toggle",
    repeat_delay = 200,
    touchpad = {
      natural_scroll = true,
      drag_lock = 1,
      tap_and_drag = true,
    },
  },
  gestures = {
    workspace_swipe_forever = true,
    workspace_swipe_create_new = true,
  },
})
hl.gesture({ fingers = 3, direction = "horizontal", action = "workspace" })

-- The JetKVM's keyboard, driven from a Mac: macOS keeps Command for itself,
-- so Super never arrives. Its left Alt (the Mac's Option, which passes
-- through) is Super on this keyboard alone; the desk keyboards are untouched.
hl.device({ name = "jetkvm-usb-emulation-device", kb_options = "altwin:swap_lalt_lwin" })
