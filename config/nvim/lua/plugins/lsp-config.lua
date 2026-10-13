-- Override LazyVim LSP defaults
return {
  "neovim/nvim-lspconfig",
  opts = {
    inlay_hints = {
      enabled = false, -- Disable inlay hints globally
    },
    servers = {
      gopls = {
        capabilities = {
          workspace = {
            didChangeWatchedFiles = {
              dynamicRegistration = true,
            },
          },
        },
      },
    },
  },
}
