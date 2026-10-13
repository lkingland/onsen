-- Register which-key prefixes
return {
  "folke/which-key.nvim",
  opts = {
    spec = {
      { ",", group = "Code", icon = "" },
      { "a", group = "around", mode = { "x", "o" } },
      { "i", group = "inside", mode = { "x", "o" } },
      { "<leader>h", group = "Harpoon", icon = "󱡀" },
    },
  },
  keys = {
    { ",", function() require("which-key").show({ keys = "," }) end, desc = "Code", mode = "n" },
  },
}
