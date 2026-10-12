-- Onsen theme loader
-- Applies all highlight groups

local M = {}

function M.setup()
  local palette = require("onsen.palette")
  local c = palette.colors

  -- Collect all highlight groups
  local highlights = {}

  -- Load group modules
  local groups = {
    require("onsen.groups.syntax"),
    require("onsen.groups.editor"),
    require("onsen.groups.treesitter"),
    require("onsen.groups.lsp"),
    require("onsen.groups.plugins"),
  }

  -- Merge all groups
  for _, group in ipairs(groups) do
    local hl = group.get(c)
    for name, opts in pairs(hl) do
      highlights[name] = opts
    end
  end

  -- Apply highlights
  for group, opts in pairs(highlights) do
    vim.api.nvim_set_hl(0, group, opts)
  end

  -- Set terminal colors
  for i, color in ipairs(c.terminal) do
    vim.g["terminal_color_" .. (i - 1)] = color
  end
end

return M
