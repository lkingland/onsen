-- Onsen colorscheme palette
-- OLED-optimized green theme with blue structure, orange data

local M = {}

M.colors = {
  -- Backgrounds
  bg = "#000000",              -- OLED black
  bg_dark = "#000000",         -- Sidebars
  bg_float = "#0A1A12",        -- Subtle green tint for floats
  bg_highlight = "#0A1A12",    -- Cursor line
  bg_popup = "#0A1A12",        -- Completion menu
  bg_search = "#8FD9A8",       -- Search highlight bg
  bg_visual = "#1A3A2A",       -- Visual selection

  -- Foregrounds
  fg = "#E0E0E0",              -- Primary text
  fg_dark = "#B0B0B0",         -- Dimmed text (operators, punctuation)
  fg_gutter = "#606060",       -- Line numbers

  -- Green family
  green = "#8FD9A8",           -- Variables, functions, constants
  green_muted = "#8FD9A8",     -- Types (italic)
  green_dark = "#008855",      -- Dark green
  green_bright = "#00FF99",    -- UI only: notifications, success, callouts
  green_mid = "#00AA66",       -- Mid green

  -- Blue (structure)
  blue = "#88AAFF",            -- Keywords (saturated periwinkle)
  blue_bright = "#99BBFF",     -- Bright blue
  blue_dark = "#2D4055",       -- Indent guides, subtle structure

  -- Cyan (reserved for special cases)
  cyan = "#77DDDD",            -- Fields, methods, member access
  cyan_bright = "#88DDDD",     -- Bright cyan

  -- Orange (warnings, annotations)
  orange = "#FF9E50",          -- Warnings, special annotations, TODO
  orange_bright = "#FFBB77",   -- Bright orange

  -- Red (errors)
  red = "#FF6B6B",             -- Errors
  red_bright = "#FF8888",      -- Bright red

  -- Neutrals
  comment = "#707070",         -- Comments
  grey = "#808080",            -- Secondary info
  white = "#FFFFFF",           -- Emphasis

  -- Semantic aliases
  none = "NONE",
}

-- Git colors
M.colors.git_add = M.colors.green
M.colors.git_change = M.colors.orange
M.colors.git_delete = M.colors.red

-- Diff backgrounds (vcsigns inline hunk view, <leader>ghd)
M.colors.diff_add_bg = "#08210F"       -- added lines (full-line tint, whisper-quiet)
M.colors.diff_add_text_bg = "#143A23"  -- added words (emphasis)
M.colors.diff_del_fg = "#C98686"       -- removed line text
M.colors.diff_del_bg = "#1B0909"       -- removed lines (virtual, whisper-quiet)
M.colors.diff_del_text_bg = "#5E1F1F"  -- removed words (emphasis)

-- Diagnostic colors
M.colors.error = M.colors.red
M.colors.warning = M.colors.orange
M.colors.info = M.colors.blue
M.colors.hint = M.colors.cyan

-- Terminal colors (ANSI 0-15)
M.colors.terminal = {
  M.colors.bg,            -- 0: black
  M.colors.red,           -- 1: red
  M.colors.green,         -- 2: green
  M.colors.orange,        -- 3: yellow/orange
  M.colors.blue,          -- 4: blue
  M.colors.green_dark,    -- 5: magenta -> dark green
  M.colors.cyan,          -- 6: cyan
  M.colors.fg_dark,       -- 7: white (dim)
  M.colors.comment,       -- 8: bright black
  M.colors.red_bright,    -- 9: bright red
  M.colors.green_bright,  -- 10: bright green
  M.colors.orange_bright, -- 11: bright yellow/orange
  M.colors.blue_bright,   -- 12: bright blue
  M.colors.green_mid,     -- 13: bright magenta -> mid green
  M.colors.cyan_bright,   -- 14: bright cyan
  M.colors.white,         -- 15: bright white
}

return M
