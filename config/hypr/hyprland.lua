-- Learn how to configure Hyprland: https://wiki.hypr.land/Configuring/Start/

-- Omarchy's bootstrap keeps path setup out of this user config.
dofile((os.getenv("OMARCHY_PATH") or "/usr/share/omarchy") .. "/default/hypr/bootstrap.lua")

-- Disable all Omarchy default bindings. Add your own in hypr/bindings.lua.
-- omarchy_default_bindings = false
--
-- Or disable only bindings for Omarchy's preinstalled apps/web apps while
-- keeping core window-manager bindings:
-- omarchy_preinstalled_bindings = false

-- Load Omarchy defaults.
require("default.hypr.omarchy")

-- Put your personal overrides in these files. They're loaded after Omarchy's
-- defaults so package updates can improve the defaults without rewriting your
-- ~/.config/hypr files.
require("hypr.monitors")
require("hypr.input")
require("hypr.bindings")
require("hypr.looknfeel")
require("hypr.autostart")

-- Toggle config flags dynamically.
require("default.hypr.toggles")

-- Add any other personal Hyprland configuration below.
-- o.window("qemu", { workspace = "5" })

-- Thunderbird's calendar popup shares the main window's class. It maps with
-- an empty title; the event name is set after, and a float rule only sees
-- the initial title. The editor maps as "Edit Item" and is retitled the
-- same way. The main window maps as "Mozilla Thunderbird" and stays tiled.
o.window({
  class = "^org\\.mozilla\\.Thunderbird$",
  initial_title = "^$|^Edit Item$|^Edit Recurrence$|^Invite Attendees$|^Please Specify the Timezone$",
}, { float = true, center = true, size = { 860, 980 } })

-- Calendar reminders map as "Calendar Reminders" and are retitled
-- "1 Reminder" once the list loads. The dialog's own size is 600 by 300.
o.window({
  class = "^org\\.mozilla\\.Thunderbird$",
  initial_title = "^Calendar Reminders$",
}, { float = true, center = true, size = { 600, 300 } })
