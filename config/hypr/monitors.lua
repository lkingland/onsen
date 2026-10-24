-- See https://wiki.hypr.land/Configuring/Basics/Monitors/
-- List current monitors and supported resolutions with: hyprctl monitors all

-- Monitor scale is Hyprland's scale for the output. It sizes everything
-- Wayland-native, accepts fractions (1.6, 1.75), and applies immediately.
-- "auto" lets Hyprland pick per display. These desks are LG 4K OLED TVs,
-- two 42" and a 48", and 1.25 is the step up from 1. At that scale each is
-- 3072 logical pixels wide, which is what the positions below count in.
local omarchy_monitor_scale = "1.25"
-- The TVs do 120 Hz over HDMI 2.1 but report 60 Hz as preferred, so the mode
-- is named.
local omarchy_monitor_mode = "3840x2160@120"
-- All three TVs report the same make, model and serial (0x01010101), so a
-- desc: match cannot tell them apart and the connector is the only stable
-- handle. HDMI-A-1 is the left desk, DP-3 the centre, DP-2 the right one
-- (the 48", added 2026-10-24). Positions are explicit so the row survives
-- any enumeration or hotplug order; auto-left only placed it right by luck.
hl.monitor({ output = "HDMI-A-1", mode = omarchy_monitor_mode, position = "0x0", scale = omarchy_monitor_scale })
hl.monitor({ output = "DP-3", mode = omarchy_monitor_mode, position = "3072x0", scale = omarchy_monitor_scale })
hl.monitor({ output = "DP-2", mode = omarchy_monitor_mode, position = "6144x0", scale = omarchy_monitor_scale })
-- Any other output that appears lands to the left of the row.
hl.monitor({ output = "", mode = omarchy_monitor_mode, position = "auto-left", scale = omarchy_monitor_scale })
-- The JetKVM: its own 1080p screen at scale 1, placed to the left like any
-- later output. Matched by description, so it follows the KVM and not the
-- port.
hl.monitor({ output = "desc:JET POWER TECHNOLOGY CO. LTD. JetKVM v1 0xC0FFEE01", mode = "1920x1080@60", position = "auto-left", scale = 1 })

-- Configure a specific monitor.
-- hl.monitor({ output = "DP-2", mode = "2560x1440@144", position = "0x0", scale = 1 })

-- Portrait/rotated secondary monitor (transform: 1 = 90°, 3 = 270°).
-- hl.monitor({ output = "DP-2", mode = "preferred", position = "auto", scale = 1, transform = 1 })

-- GDK scale is GDK_SCALE, the factor GTK draws its own UI at. It's what
-- sizes X11/XWayland windows, which Omarchy leaves unscaled so they stay
-- crisp instead of being stretched by the compositor. GTK only honors whole
-- numbers, so use the nearest integer to the monitor scale, and restart an
-- app for a change to reach it.
local omarchy_gdk_scale = 2
hl.env("GDK_SCALE", tostring(omarchy_gdk_scale))
