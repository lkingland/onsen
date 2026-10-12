-- Onsen syntax highlight groups
-- Base Vim syntax groups

local M = {}

function M.get(c)
  return {
    -- Comments
    Comment = { fg = c.comment },

    -- Constants
    Constant = { fg = c.fg },
    String = { fg = c.cyan },
    Character = { fg = c.cyan },
    Number = { fg = c.fg },
    Boolean = { fg = c.fg },
    Float = { fg = c.fg },

    -- Identifiers
    Identifier = { fg = c.fg },
    Function = { fg = c.fg },

    -- Statements (blue: control flow + reserved words)
    Statement = { fg = c.blue },
    Conditional = { fg = c.blue },
    Repeat = { fg = c.blue },
    Label = { fg = c.blue },
    Operator = { fg = c.fg_dark },
    Keyword = { fg = c.blue },
    Exception = { fg = c.blue },

    -- Preprocessor
    PreProc = { fg = c.blue },
    Include = { fg = c.blue },
    Define = { fg = c.blue },
    Macro = { fg = c.blue },
    PreCondit = { fg = c.blue },

    -- Types (italic only, no color accent)
    Type = { fg = c.fg, italic = true },
    StorageClass = { fg = c.blue },
    Structure = { fg = c.fg, italic = true },
    Typedef = { fg = c.fg, italic = true },

    -- Special
    Special = { fg = c.orange },
    SpecialChar = { fg = c.orange },
    Tag = { fg = c.green },
    Delimiter = { fg = c.grey },
    SpecialComment = { fg = c.orange },
    Debug = { fg = c.orange },

    -- Underlined
    Underlined = { underline = true },

    -- Ignore
    Ignore = { fg = c.comment },

    -- Error
    Error = { fg = c.red },

    -- Todo
    Todo = { fg = c.bg, bg = c.orange, bold = true },

    -- Go-specific
    goBlock = { fg = c.grey },
    goParen = { fg = c.grey },
    goBraces = { fg = c.grey },
  }
end

return M
