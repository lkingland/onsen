-- fff.nvim: Rust-powered fuzzy file finder with frecency scoring
-- Replaces Snacks picker for file finding and grep only
--
-- Disabled 2026-05-14: hangs nvim startup when multiple Claude Code sessions
-- are concurrently running fff-mcp. Both share ~/.cache/nvim/fff_nvim/ LMDB,
-- and fff.nvim deadlocks waiting on locks held by the MCP daemons. Falls back
-- to LazyVim's Snacks picker (default `<leader>ff`, `<leader>fg`, etc.).
return {
  "dmtrKovalenko/fff.nvim",
  enabled = false,
  build = function()
    require("fff.download").download_or_build_binary()
  end,
  lazy = false,
  keys = {
    { "<leader>ff", function() require("fff").find_files() end, desc = "Find files (fff)" },
    { "<leader><space>", function() require("fff").find_files() end, desc = "Find files (fff)" },
    { "<leader>fF", function() require("fff").find_files_in_dir(vim.fn.expand("%:p:h")) end, desc = "Find files (cwd, fff)" },
    { "<leader>fg", function() require("fff").live_grep() end, desc = "Live grep (fff)" },
    { "<leader>/", function() require("fff").live_grep() end, desc = "Grep (fff)" },
  },
  opts = {},
}
