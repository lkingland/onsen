-- Onsen editor highlight groups
-- UI elements: windows, menus, statusline, etc.

local M = {}

function M.get(c)
  return {
    -- Normal text
    Normal = { fg = c.fg, bg = c.bg },
    NormalNC = { fg = c.fg, bg = c.bg },
    NormalFloat = { fg = c.fg, bg = c.bg_float },
    NormalSB = { fg = c.fg, bg = c.bg_dark },

    -- Cursor
    Cursor = { fg = c.bg, bg = c.fg },
    lCursor = { fg = c.bg, bg = c.fg },
    CursorIM = { fg = c.bg, bg = c.fg },
    CursorLine = { bg = c.bg_highlight },
    CursorColumn = { bg = c.bg_highlight },
    CursorLineNr = { fg = c.green, bold = true },

    -- Line numbers
    LineNr = { fg = c.fg_gutter },
    LineNrAbove = { fg = c.fg_gutter },
    LineNrBelow = { fg = c.fg_gutter },
    SignColumn = { fg = c.fg_gutter, bg = c.bg },
    FoldColumn = { fg = c.comment, bg = c.bg },

    -- Separators
    VertSplit = { fg = c.comment },
    WinSeparator = { fg = c.comment },
    ColorColumn = { bg = "#1A2A22" },

    -- Popup menu
    Pmenu = { fg = c.fg, bg = c.bg_popup },
    PmenuSel = { fg = c.bg, bg = c.green },
    PmenuSbar = { bg = c.bg_popup },
    PmenuThumb = { bg = c.green_dark },
    PmenuMatch = { fg = c.orange, bg = c.bg_popup },
    PmenuMatchSel = { fg = c.orange, bg = c.green },

    -- Floating windows
    FloatBorder = { fg = c.green, bg = c.bg_float },
    FloatTitle = { fg = c.green, bg = c.bg_float },
    FloatFooter = { fg = c.comment, bg = c.bg_float },

    -- Search
    Search = { fg = c.bg, bg = c.bg_search },
    IncSearch = { fg = c.bg, bg = c.orange },
    CurSearch = { fg = c.bg, bg = c.orange },
    Substitute = { fg = c.bg, bg = c.red },

    -- Visual
    Visual = { bg = c.bg_visual },
    VisualNOS = { bg = c.bg_visual },

    -- Folding
    Folded = { fg = c.comment, bg = c.bg_highlight },

    -- Status line
    StatusLine = { fg = c.fg, bg = c.bg_dark },
    StatusLineNC = { fg = c.comment, bg = c.bg_dark },
    WinBar = { fg = c.fg, bg = c.bg },
    WinBarNC = { fg = c.comment, bg = c.bg },

    -- Tab line
    TabLine = { fg = c.comment, bg = c.bg_dark },
    TabLineFill = { bg = c.bg_dark },
    TabLineSel = { fg = c.fg, bg = c.bg },

    -- Messages
    ErrorMsg = { fg = c.red },
    WarningMsg = { fg = c.orange },
    MoreMsg = { fg = c.green },
    ModeMsg = { fg = c.fg, bold = true },
    MsgArea = { fg = c.fg },
    Question = { fg = c.blue },
    Title = { fg = c.green, bold = true },

    -- Diff
    DiffAdd = { fg = c.git_add, bg = c.bg_highlight },
    DiffChange = { fg = c.git_change, bg = c.bg_highlight },
    DiffDelete = { fg = c.git_delete, bg = c.bg_highlight },
    DiffText = { fg = c.bg, bg = c.git_change },

    -- Spelling
    SpellBad = { sp = c.red, undercurl = true },
    SpellCap = { sp = c.orange, undercurl = true },
    SpellLocal = { sp = c.cyan, undercurl = true },
    SpellRare = { sp = c.blue, undercurl = true },

    -- Misc
    Directory = { fg = c.green },
    MatchParen = { fg = c.orange, bold = true },
    NonText = { fg = c.comment },
    SpecialKey = { fg = c.comment },
    Whitespace = { fg = c.comment },
    EndOfBuffer = { fg = c.bg },
    WildMenu = { fg = c.bg, bg = c.green },
    Conceal = { fg = c.comment },
    QuickFixLine = { bg = c.bg_visual },

    -- Health
    healthError = { fg = c.red },
    healthSuccess = { fg = c.green },
    healthWarning = { fg = c.orange },
  }
end

return M
