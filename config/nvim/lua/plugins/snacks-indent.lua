-- Customize snacks.nvim indent guides
return {
  "folke/snacks.nvim",
  opts = {
    indent = {
      enabled = true,
      char = "│",
    },
    scope = {
      enabled = true,
      char = "┃",  -- heavy vertical (thicker than │)
    },
  },
}
