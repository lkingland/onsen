-- Onsen's keymaps. LazyVim sets its own on VeryLazy, after plugin/ has run,
-- so these wait for that event too and win where they overlap: Alt+hjkl
-- resizes here, where LazyVim moves lines. lua/config/keymaps.lua stays as
-- the starter ships it.
vim.api.nvim_create_autocmd("User", {
  pattern = "VeryLazy",
  callback = function()
    local map = vim.keymap.set

    -- Quit
    map({ "n", "i", "v" }, "<C-q>", "<cmd>q<cr>", { desc = "Quit" })
    map("n", "QQ", "<cmd>q!<cr>", { desc = "Force Quit" })

    -- Buffer and window navigation
    map({ "n", "i", "v" }, "<S-Tab>", "<C-^>", { desc = "Toggle Previous Buffer", silent = true })
    map({ "n", "i", "v" }, "<A-Tab>", "<cmd>wincmd p<cr>", { desc = "Toggle Previous Window", silent = true })

    -- Window resizing
    map("n", "<A-h>", "<cmd>vertical resize -2<cr>", { desc = "Shrink Window Width" })
    map("n", "<A-l>", "<cmd>vertical resize +2<cr>", { desc = "Expand Window Width" })
    map("n", "<A-k>", "<cmd>resize -2<cr>", { desc = "Shrink Window Height" })
    map("n", "<A-j>", "<cmd>resize +2<cr>", { desc = "Expand Window Height" })

    -- Yank the whole buffer in place. yy is the line; :%y and ggyG move or
    -- take a command. unnamedplus already copies to the system clipboard.
    map("n", "yY", "<cmd>%y<cr>", { desc = "Yank entire buffer" })
  end,
})
