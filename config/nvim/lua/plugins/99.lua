-- ThePrimeagen's 99: targeted AI code replacement
-- Use <leader>9 prefix for all 99 commands
return {
  "ThePrimeagen/99",
  keys = {
    { "<leader>9v", function() require("99").visual() end, mode = "v", desc = "99 Visual" },
    { "<leader>9s", function() require("99").stop_all_requests() end, desc = "99 Stop All" },
  },
  config = function()
    local _99 = require("99")
    _99.setup({
      provider = _99.Providers.ClaudeCodeProvider,
      model = "claude-sonnet-4-6",
    })
    require("which-key").add({
      { "<leader>9", group = "+99", icon = "" },
    })
  end,
}
