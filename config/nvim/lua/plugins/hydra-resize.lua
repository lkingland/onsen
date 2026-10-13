-- Resize mode: Ctrl-w then h/j/k/l to resize, Esc or any other key to exit
return {
  "nvimtools/hydra.nvim",
  keys = {
    { "<C-w>", desc = "Window resize mode" },
  },
  config = function()
    local Hydra = require("hydra")

    Hydra({
      name = "Resize Window",
      mode = "n",
      body = "<C-w>",
      hint = [[
 Resize: _h_ ← _l_ →  _k_ ↑ _j_ ↓
 _=_ equalize  _q_/Esc exit
]],
      config = {
        color = "pink",
        invoke_on_body = true,
        hint = {
          type = "window",
          position = "bottom",
          border = "rounded",
        },
      },
      heads = {
        { "h", "<cmd>vertical resize -2<cr>", { desc = "Shrink width" } },
        { "l", "<cmd>vertical resize +2<cr>", { desc = "Expand width" } },
        { "k", "<cmd>resize -2<cr>", { desc = "Shrink height" } },
        { "j", "<cmd>resize +2<cr>", { desc = "Expand height" } },
        { "=", "<C-w>=", { desc = "Equalize" } },
        -- Window navigation (pass-through to default Ctrl-w behavior)
        { "H", "<C-w>H", { exit = true, desc = "Move window left" } },
        { "J", "<C-w>J", { exit = true, desc = "Move window down" } },
        { "K", "<C-w>K", { exit = true, desc = "Move window up" } },
        { "L", "<C-w>L", { exit = true, desc = "Move window right" } },
        { "s", "<C-w>s", { exit = true, desc = "Split horizontal" } },
        { "v", "<C-w>v", { exit = true, desc = "Split vertical" } },
        { "c", "<C-w>c", { exit = true, desc = "Close window" } },
        { "o", "<C-w>o", { exit = true, desc = "Only this window" } },
        { "q", nil, { exit = true, desc = "Exit" } },
        { "<Esc>", nil, { exit = true, desc = false } },
      },
    })
  end,
}
