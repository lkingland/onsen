-- Configure Trouble to show symbols on the left with custom width
return {
  "folke/trouble.nvim",
  opts = function(_, opts)
    opts.modes = opts.modes or {}
    opts.modes.symbols = opts.modes.symbols or {}
    opts.modes.symbols.auto_preview = false
    opts.modes.symbols.win = {
      type = "split",
      relative = "win",
      position = "left",
      size = 40,
    }
    opts.modes.lsp = opts.modes.lsp or {}
    opts.modes.lsp.win = {
      type = "split",
      relative = "win",
      position = "left",
      size = 40,
    }
    return opts
  end,
}
