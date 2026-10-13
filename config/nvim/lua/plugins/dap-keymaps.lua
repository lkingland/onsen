-- Debug keymaps using comma prefix (,d)
-- Extends the DAP core extra with custom keybindings
return {
  "mfussenegger/nvim-dap",
  keys = {
    { ",db", function() require("dap").toggle_breakpoint() end, desc = "Toggle Breakpoint" },
    { ",dc", function() require("dap").continue() end, desc = "Continue" },
    { ",di", function() require("dap").step_into() end, desc = "Step Into" },
    { ",do", function() require("dap").step_over() end, desc = "Step Over" },
    { ",du", function() require("dap").step_out() end, desc = "Step Out" },
    { ",dl", function() require("dap").run_last() end, desc = "Run Last" },
    { ",dx", function() require("dap").terminate() end, desc = "Terminate" },
  },
  config = function()
    require("which-key").add({
      { ",d", group = "+Debug", icon = "" },
    })
  end,
}
