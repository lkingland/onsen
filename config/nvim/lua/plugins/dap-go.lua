-- Go-specific debugging with delve
-- Adds ,dt for debugging the current test
return {
  "leoluz/nvim-dap-go",
  ft = "go",
  dependencies = { "mfussenegger/nvim-dap" },
  opts = {},
  keys = {
    { ",dt", function() require("dap-go").debug_test() end, desc = "Debug Test", ft = "go" },
    { ",dr", function() require("dap-go").debug_last_test() end, desc = "Debug Last Test", ft = "go" },
  },
}
