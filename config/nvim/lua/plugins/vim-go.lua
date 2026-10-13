-- Full Go development environment with vim-go
return {
  "fatih/vim-go",
  ft = "go",
  init = function()
    -- Disable vim-go's LSP, let LazyVim's gopls handle it
    vim.g.go_gopls_enabled = 0
    vim.g.go_def_mapping_enabled = 0  -- Let LSP handle go-to-definition

    vim.g.go_test_show_name = 1
    vim.g.go_test_timeout = "20s"
    vim.g.go_highlight_types = 1
    vim.g.go_highlight_fields = 1
    vim.g.go_highlight_functions = 1
    vim.g.go_highlight_function_calls = 1
    vim.g.go_highlight_operators = 1
    vim.g.go_highlight_extra_types = 1
    vim.g.go_highlight_build_constraints = 1
    vim.g.go_highlight_generate_tags = 1
  end,
  config = function()
    require("which-key").add({
      { ",", group = "+Code", icon = "" },
      { ",s", ":GoAlternate<cr>", desc = "Switch to Alternate" },
      { ",b", ":w<bar>GoBuild<cr>", desc = "Build", silent = true },
      { ",r", ":GoRun<cr>", desc = "Run" },
      { ",c", group = "+Coverage" },
      { ",cc", ":GoCoverageToggle<cr>", desc = "Coverage toggle" },
      { ",cb", ":GoCoverageBrowser<cr>", desc = "Browse coverage" },
      { ",t", group = "+Test" },
      { ",tt", ":w<bar>:GoTest<cr>", desc = "Test all", silent = true },
      { ",tf", ":w<bar>:GoTestFunc<cr>", desc = "Test function", silent = true },
      { ",ti", ":w<bar>:GoTestFile<cr>", desc = "Test file", silent = true },
      { ",tc", ":w<bar>:GoTestCompile<cr>", desc = "Test compile", silent = true },
      { ",tm", ":w<bar>:GoMetaLinter<cr>", desc = "Metalinter", silent = true },
      { ",x", group = "+Docs" },
      { ",xi", ":GoDoc<cr>", desc = "Inline docs" },
      { ",xb", ":GoDocBrowser<cr>", desc = "Browser docs" },
    })
  end,
}
