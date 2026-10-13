-- Render markdown with inline formatting (headings, tables, checkboxes, etc.)
return {
  "MeanderingProgrammer/render-markdown.nvim",
  dependencies = { "nvim-treesitter/nvim-treesitter" },
  ft = { "markdown" },
  opts = {
    bullet = { right_pad = 1 },
    code = { highlight = "", highlight_border = "", highlight_language = "" },
  },
  init = function()
    -- conceallevel and wiki link navigation for markdown
    vim.api.nvim_create_autocmd("FileType", {
      pattern = "markdown",
      callback = function()
        vim.wo.conceallevel = 2
        vim.opt_local.suffixesadd:append(".md")

        -- gf override: resolve [[Wiki Link]] as wiki-link.md
        vim.keymap.set("n", "gf", function()
          local line = vim.api.nvim_get_current_line()
          local col = vim.api.nvim_win_get_cursor(0)[2] + 1
          -- Check if cursor is inside [[ ]]
          local s, e, link = line:find("%[%[(.-)%]%]")
          while s do
            if col >= s and col <= e then
              local dir = vim.fn.expand("%:p:h")
              local file = dir .. "/" .. link:lower():gsub("%s+", "-") .. ".md"
              vim.cmd("edit " .. vim.fn.fnameescape(file))
              return
            end
            s, e, link = line:find("%[%[(.-)%]%]", e + 1)
          end
          -- Fall back to default gf
          vim.cmd("normal! gF")
        end, { buffer = true, desc = "Follow markdown link" })
      end,
    })
  end,
}
