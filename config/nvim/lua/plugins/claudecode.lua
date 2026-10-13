-- Claude Code integration for AI-assisted coding
-- Use <leader>a prefix for all Claude commands
return {
  "coder/claudecode.nvim",
  dependencies = { "folke/snacks.nvim" },
  opts = {
    terminal = {
      split_side = "right",
      split_width_percentage = 0.30,
    },
    terminal_cmd = "claude --ide --model opus[1m] --append-system-prompt-file ~/dotfiles/claude-nvim-preamble.md",
    models = {
      { name = "Claude Opus 4.7 (1M)", value = "opus[1m]" },
      { name = "Claude Opus 4.7", value = "opus" },
      { name = "Claude Sonnet 4.6", value = "sonnet" },
      { name = "Claude Haiku 4.5", value = "haiku" },
    },
  },
  keys = {
    { "<leader>a", nil, desc = "AI/Claude Code" },
    { "<leader>ac", "<cmd>ClaudeCode<cr>", desc = "Toggle Claude" },
    { "<leader>af", "<cmd>ClaudeCodeFocus<cr>", desc = "Focus Claude" },
    { "<leader>ar", "<cmd>ClaudeCode --resume<cr>", desc = "Resume Claude" },
    { "<leader>aC", "<cmd>ClaudeCode --continue<cr>", desc = "Continue Claude" },
    { "<leader>am", "<cmd>ClaudeCodeSelectModel<cr>", desc = "Select Claude model" },
    { "<leader>ab", "<cmd>ClaudeCodeAdd %<cr>", desc = "Add current buffer" },
    { "<leader>as", function()
        vim.cmd("ClaudeCodeSend")
        vim.schedule(function()
          vim.cmd("ClaudeCodeFocus")
          vim.cmd("startinsert")
        end)
      end, mode = "v", desc = "Send to Claude" },
    { "<leader>aa", "<cmd>ClaudeCodeDiffAccept<cr>", desc = "Accept diff" },
    { "<leader>ad", "<cmd>ClaudeCodeDiffDeny<cr>", desc = "Deny diff" },
    { "<leader>aR", function()
        vim.cmd("checktime") -- reload unmodified buffers whose files changed on disk
        if not vim.bo.modified then
          pcall(vim.cmd.edit)
        end
        -- nvim 0.12 ships :lsp; nvim-lspconfig then skips defining :LspRestart
        if vim.fn.exists(":lsp") == 2 then
          vim.cmd("lsp restart")
        elseif vim.fn.exists(":LspRestart") == 2 then
          vim.cmd("LspRestart")
        end
        local ok, todo = pcall(require, "todo-comments")
        if ok and todo.refresh then
          todo.refresh()
        end
      end, desc = "Refresh buffers & LSP" },
  },
  config = function(_, opts)
    require("claudecode").setup(opts)
    require("which-key").add({
      { "<leader>a", group = "+AI/Claude", icon = "󰚩" },
    })
  end,
}
