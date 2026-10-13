-- Customize todo-comments highlighting
-- Only highlight the keyword itself, not the full comment
return {
  "folke/todo-comments.nvim",
  opts = {
    highlight = {
      before = "",           -- nothing before keyword
      keyword = "bg",        -- highlight just the keyword with background
      after = "",            -- don't color the text after keyword
      comments_only = true,  -- only highlight in comments (via treesitter)
    },
  },
}
