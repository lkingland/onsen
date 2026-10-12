-- Onsen colorscheme
-- OLED-optimized green theme for neovim

local M = {}

function M.load()
  -- Clear existing highlights
  if vim.g.colors_name then
    vim.cmd("hi clear")
  end

  -- Enable true colors
  vim.o.termguicolors = true
  vim.o.background = "dark"

  -- Set colorscheme name
  vim.g.colors_name = "onsen"

  -- Load theme
  require("onsen.theme").setup()
end

-- Convenience setup function (for lazy.nvim compatibility)
function M.setup(opts)
  -- Currently no options, but structure allows for future expansion
  M.load()
end

return M
