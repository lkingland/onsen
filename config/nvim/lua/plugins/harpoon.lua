-- Quick file marking and navigation
-- Use <leader>ha to add files, Alt+1-9 to jump to them
return {
  "ThePrimeagen/harpoon",
  branch = "harpoon2",
  dependencies = { "nvim-lua/plenary.nvim" },
  opts = {},
  keys = {
    { "<leader>ha", function() require("harpoon"):list():add() end, desc = "Harpoon Add" },
    { "<leader>hh", function() require("harpoon").ui:toggle_quick_menu(require("harpoon"):list()) end, desc = "Harpoon Menu" },
    { "<leader>hp", function() require("harpoon"):list():prev() end, desc = "Harpoon Prev" },
    { "<leader>hn", function() require("harpoon"):list():next() end, desc = "Harpoon Next" },
    { "<A-1>", function() require("harpoon"):list():select(1) end, desc = "Harpoon 1" },
    { "<A-2>", function() require("harpoon"):list():select(2) end, desc = "Harpoon 2" },
    { "<A-3>", function() require("harpoon"):list():select(3) end, desc = "Harpoon 3" },
    { "<A-4>", function() require("harpoon"):list():select(4) end, desc = "Harpoon 4" },
    { "<A-5>", function() require("harpoon"):list():select(5) end, desc = "Harpoon 5" },
    { "<A-6>", function() require("harpoon"):list():select(6) end, desc = "Harpoon 6" },
    { "<A-7>", function() require("harpoon"):list():select(7) end, desc = "Harpoon 7" },
    { "<A-8>", function() require("harpoon"):list():select(8) end, desc = "Harpoon 8" },
    { "<A-9>", function() require("harpoon"):list():select(9) end, desc = "Harpoon 9" },
  },
  config = function(_, opts)
    require("harpoon"):setup(opts)
    require("which-key").add({
      { "<leader>h", group = "+Harpoon", icon = "󱡀" },
    })
  end,
}
