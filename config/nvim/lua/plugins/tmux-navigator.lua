-- Seamless navigation between tmux panes and vim windows
return {
  "christoomey/vim-tmux-navigator",
  event = "VeryLazy",
  init = function()
    vim.g.tmux_navigator_no_wrap = 1
  end,
  cmd = {
    "TmuxNavigateLeft",
    "TmuxNavigateDown",
    "TmuxNavigateUp",
    "TmuxNavigateRight",
    "TmuxNavigatePrevious",
  },
  keys = {
    -- Normal mode: <cmd> works fine
    { "<C-h>", "<cmd>TmuxNavigateLeft<cr>", mode = "n", desc = "Navigate Left" },
    { "<C-j>", "<cmd>TmuxNavigateDown<cr>", mode = "n", desc = "Navigate Down" },
    { "<C-k>", "<cmd>TmuxNavigateUp<cr>", mode = "n", desc = "Navigate Up" },
    { "<C-l>", "<cmd>TmuxNavigateRight<cr>", mode = "n", desc = "Navigate Right" },
    { "<C-\\>", "<cmd>TmuxNavigatePrevious<cr>", mode = "n", desc = "Navigate Previous" },
    -- Terminal mode: must escape with <C-\><C-n> first so tmux passthrough works
    { "<C-h>", "<C-\\><C-n><cmd>TmuxNavigateLeft<cr>", mode = "t", desc = "Navigate Left" },
    { "<C-j>", "<C-\\><C-n><cmd>TmuxNavigateDown<cr>", mode = "t", desc = "Navigate Down" },
    { "<C-k>", "<C-\\><C-n><cmd>TmuxNavigateUp<cr>", mode = "t", desc = "Navigate Up" },
    { "<C-l>", "<C-\\><C-n><cmd>TmuxNavigateRight<cr>", mode = "t", desc = "Navigate Right" },
    { "<C-\\>", "<C-\\><C-n><cmd>TmuxNavigatePrevious<cr>", mode = "t", desc = "Navigate Previous" },
  },
}
