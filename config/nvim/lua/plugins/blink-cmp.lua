-- Override blink.cmp to be opt-in (manual trigger) like legacy nvim-cmp setup
return {
  "saghen/blink.cmp",
  opts = {
    completion = {
      trigger = {
        show_on_keyword = false,
        show_on_trigger_character = false,
        show_on_insert_on_trigger_character = false,
        show_on_accept_on_trigger_character = false,
      },
      menu = {
        auto_show = false,
      },
      documentation = {
        auto_show = true,
        auto_show_delay_ms = 200,
      },
    },
    keymap = {
      preset = "none",
      ["<C-space>"] = { "show", "show_documentation", "hide_documentation" },
      ["<C-a>"] = { "show", "show_documentation", "hide_documentation" },
      ["<C-y>"] = { "select_and_accept" },
      ["<CR>"] = { "accept", "fallback" },
      ["<C-n>"] = { "select_next", "fallback" },
      ["<C-p>"] = { "select_prev", "fallback" },
      ["<C-b>"] = { "scroll_documentation_up", "fallback" },
      ["<C-f>"] = { "scroll_documentation_down", "fallback" },
      ["<C-l>"] = { "snippet_forward", "fallback" },
      ["<C-L>"] = { "snippet_backward", "fallback" },
    },
  },
}
