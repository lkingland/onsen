-- Beancount accounting format support
return {
  "nathangrigg/vim-beancount",
  ft = { "beancount", "bean" },
  init = function()
    -- Set beancount root for account completion
    vim.api.nvim_create_autocmd("FileType", {
      pattern = "beancount",
      callback = function()
        vim.b.beancount_root = vim.fn.expand("~/src/kingland.io/pta/main.bean")
      end,
    })
  end,
}
