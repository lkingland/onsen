-- Onsen's editor options, on top of LazyVim's and Omarchy's. Neovim runs
-- plugin/ after init.lua, so these land after lua/config/options.lua, which
-- stays as Omarchy ships it. The clipboard is Omarchy's remote_clipboard.

-- Line numbers
vim.opt.relativenumber = true
vim.opt.numberwidth = 2

-- Format on save: LazyVim's default, which Omarchy turns off
vim.g.autoformat = true

-- Visual guides
vim.opt.colorcolumn = "80"
vim.opt.scrolloff = 8
vim.opt.fillchars = { eob = " " } -- no ~ on empty lines
vim.opt.conceallevel = 0 -- do not hide characters

-- Indentation: 4 spaces
vim.opt.tabstop = 4
vim.opt.shiftwidth = 4
vim.opt.softtabstop = 4

-- Undo history
vim.opt.undolevels = 10000
vim.opt.undoreload = 100000

-- Comment continuation
vim.opt.formatoptions:append("r") -- on Enter
vim.opt.formatoptions:append("o") -- on o and O

-- Cursor: blinking block in normal, line in insert
vim.opt.guicursor = "n-v-c:block,i-ci-ve:ver25,r-cr:hor20,o:hor50"
  .. ",a:blinkwait700-blinkoff400-blinkon250-Cursor/lCursor"
  .. ",sm:block-blinkwait175-blinkoff150-blinkon175"
