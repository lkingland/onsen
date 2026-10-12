-- Lualine in the Onsen palette, with rounded pill ends. Replaces the theme
-- Omarchy's lualine-transparent.lua builds; file order puts this spec after it.
return {
  "nvim-lualine/lualine.nvim",
  opts = function(_, opts)
    local colors = {
      bg = "#000000",
      fg = "#E0E0E0",
      green_dark = "#2D5540",    -- Dark green (matches tmux active pill)
      green_darker = "#1E3A2D",  -- Darker green (matches tmux inactive pill)
      red_dark = "#552D2D",      -- Dark red
      orange_dark = "#554020",   -- Dark orange
      comment = "#606060",
    }

    local onsen_theme = {
      normal = {
        a = { fg = colors.fg, bg = colors.green_dark, gui = "bold" },
        b = { fg = colors.fg, bg = colors.green_darker },
        c = { fg = colors.fg, bg = colors.bg },
      },
      insert = {
        a = { fg = colors.fg, bg = colors.green_dark, gui = "bold" },
      },
      visual = {
        a = { fg = colors.fg, bg = colors.orange_dark, gui = "bold" },
      },
      replace = {
        a = { fg = colors.fg, bg = colors.red_dark, gui = "bold" },
      },
      command = {
        a = { fg = colors.fg, bg = colors.green_dark, gui = "bold" },
      },
      inactive = {
        a = { fg = colors.comment, bg = colors.bg },
        b = { fg = colors.comment, bg = colors.bg },
        c = { fg = colors.comment, bg = colors.bg },
      },
    }

    opts.options = opts.options or {}
    opts.options.theme = onsen_theme

    -- Left end cap (rounded pill edge)
    local left_cap = {
      function() return "" end,
      color = { fg = colors.green_dark, bg = colors.bg },
      padding = 0,
    }

    -- Right end cap (rounded pill edge)
    local right_cap = {
      function() return "" end,
      color = { fg = colors.green_dark, bg = colors.bg },
      padding = 0,
    }

    -- Insert caps at outer edges
    opts.sections = opts.sections or {}
    opts.sections.lualine_a = opts.sections.lualine_a or {}
    opts.sections.lualine_z = opts.sections.lualine_z or {}
    table.insert(opts.sections.lualine_a, 1, left_cap)
    table.insert(opts.sections.lualine_z, right_cap)
  end,
}
