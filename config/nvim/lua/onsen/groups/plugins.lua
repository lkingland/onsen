-- Onsen plugin highlight groups
-- Plugin-specific highlights (can be extended as needed)

local M = {}

function M.get(c)
  return {
    -- Telescope
    TelescopeBorder = { fg = c.green, bg = c.bg_float },
    TelescopeNormal = { fg = c.fg, bg = c.bg_float },
    TelescopePromptBorder = { fg = c.green, bg = c.bg_float },
    TelescopePromptNormal = { fg = c.fg, bg = c.bg_float },
    TelescopePromptPrefix = { fg = c.green },
    TelescopePromptTitle = { fg = c.bg, bg = c.green },
    TelescopePreviewBorder = { fg = c.green, bg = c.bg_float },
    TelescopePreviewNormal = { bg = c.bg_float },
    TelescopePreviewTitle = { fg = c.bg, bg = c.green },
    TelescopeResultsBorder = { fg = c.green, bg = c.bg_float },
    TelescopeResultsNormal = { bg = c.bg_float },
    TelescopeResultsTitle = { fg = c.bg, bg = c.green },
    TelescopeSelection = { fg = c.fg, bg = c.bg_visual },
    TelescopeSelectionCaret = { fg = c.green },
    TelescopeMatching = { fg = c.orange, bold = true },

    -- Blink.cmp (completion)
    BlinkCmpMenu = { fg = c.fg, bg = c.bg_popup },
    BlinkCmpMenuBorder = { fg = c.green, bg = c.bg_popup },
    BlinkCmpMenuSelection = { fg = c.bg, bg = c.green },
    BlinkCmpLabel = { fg = c.fg },
    BlinkCmpLabelMatch = { fg = c.orange, bold = true },
    BlinkCmpLabelDeprecated = { fg = c.comment, strikethrough = true },
    BlinkCmpKind = { fg = c.cyan },
    BlinkCmpKindText = { fg = c.fg },
    BlinkCmpKindMethod = { fg = c.green },
    BlinkCmpKindFunction = { fg = c.green },
    BlinkCmpKindConstructor = { fg = c.green_muted },
    BlinkCmpKindField = { fg = c.fg },
    BlinkCmpKindVariable = { fg = c.fg },
    BlinkCmpKindClass = { fg = c.green_muted },
    BlinkCmpKindInterface = { fg = c.green_muted },
    BlinkCmpKindModule = { fg = c.fg },
    BlinkCmpKindProperty = { fg = c.fg },
    BlinkCmpKindUnit = { fg = c.green },
    BlinkCmpKindValue = { fg = c.green },
    BlinkCmpKindEnum = { fg = c.green_muted },
    BlinkCmpKindKeyword = { fg = c.blue },
    BlinkCmpKindSnippet = { fg = c.orange },
    BlinkCmpKindColor = { fg = c.orange },
    BlinkCmpKindFile = { fg = c.fg },
    BlinkCmpKindReference = { fg = c.blue },
    BlinkCmpKindFolder = { fg = c.green },
    BlinkCmpKindEnumMember = { fg = c.green },
    BlinkCmpKindConstant = { fg = c.green },
    BlinkCmpKindStruct = { fg = c.green_muted },
    BlinkCmpKindEvent = { fg = c.orange },
    BlinkCmpKindOperator = { fg = c.fg_dark },
    BlinkCmpKindTypeParameter = { fg = c.green_muted },

    -- GitSigns (kept as fallback for vcsigns and other plugins)
    GitSignsAdd = { fg = c.git_add },
    GitSignsChange = { fg = c.git_change },
    GitSignsDelete = { fg = c.git_delete },
    GitSignsCurrentLineBlame = { fg = c.comment },

    -- VCSigns (gutter signs)
    SignAdd = { fg = c.git_add },
    SignChange = { fg = c.git_change },
    SignDelete = { fg = c.git_delete },
    SignCombined = { fg = c.git_change },

    -- VCSigns inline diff (<leader>ghd). Without these, the word-level groups
    -- fall back to DiffText (orange) for BOTH adds and deletes — an unreadable
    -- wash. Define them so the hunk reads green=added / red=removed, with the
    -- *Text* groups layering brighter emphasis over the full-line groups.
    VcsignsDiffAdd = { bg = c.diff_add_bg }, -- added lines (bg only; syntax shows)
    VcsignsDiffTextAdd = { bg = c.diff_add_text_bg, bold = true }, -- added words
    VcsignsDiffDelete = { fg = c.diff_del_fg, bg = c.diff_del_bg }, -- removed lines
    VcsignsDiffTextDelete = { fg = c.red_bright, bg = c.diff_del_text_bg, bold = true }, -- removed words

    -- Indent Blankline
    IblIndent = { fg = c.comment },
    IblScope = { fg = c.blue_dark },

    -- Which-key
    WhichKey = { fg = c.green },
    WhichKeyGroup = { fg = c.blue },
    WhichKeyDesc = { fg = c.fg },
    WhichKeySeperator = { fg = c.comment },
    WhichKeySeparator = { fg = c.comment },
    WhichKeyFloat = { bg = c.bg_float },
    WhichKeyBorder = { fg = c.green, bg = c.bg_float },
    WhichKeyValue = { fg = c.comment },

    -- Lazy.nvim
    LazyProgressDone = { fg = c.green },
    LazyProgressTodo = { fg = c.comment },

    -- Snacks.nvim
    SnacksIndent = { fg = c.blue_dark },
    SnacksIndentScope = { fg = c.comment },
    SnacksDashboardHeader = { fg = c.green },
    SnacksDashboardFooter = { fg = c.comment },
    SnacksDashboardIcon = { fg = c.green },
    SnacksDashboardDesc = { fg = c.fg },
    SnacksDashboardKey = { fg = c.orange },
    SnacksDashboardDir = { fg = c.comment },
    SnacksDashboardFile = { fg = c.fg },
    SnacksDashboardSpecial = { fg = c.cyan },
    SnacksNotifierInfo = { fg = c.info },
    SnacksNotifierWarn = { fg = c.warning },
    SnacksNotifierError = { fg = c.error },
    SnacksNotifierDebug = { fg = c.comment },
    SnacksNotifierTrace = { fg = c.comment },

    -- Neo-tree
    NeoTreeNormal = { fg = c.fg, bg = c.bg_dark },
    NeoTreeNormalNC = { fg = c.fg, bg = c.bg_dark },
    NeoTreeDirectoryName = { fg = c.blue },
    NeoTreeDirectoryIcon = { fg = c.blue },
    NeoTreeRootName = { fg = c.blue, bold = true },
    NeoTreeFileName = { fg = c.fg },
    NeoTreeFileIcon = { fg = c.fg },
    NeoTreeGitAdded = { fg = c.git_add },
    NeoTreeGitConflict = { fg = c.red },
    NeoTreeGitDeleted = { fg = c.git_delete },
    NeoTreeGitIgnored = { fg = c.comment },
    NeoTreeGitModified = { fg = c.git_change },
    NeoTreeGitUnstaged = { fg = c.git_change },
    NeoTreeGitUntracked = { fg = c.green_bright },
    NeoTreeGitStaged = { fg = c.git_add },
    NeoTreeIndentMarker = { fg = c.comment },
    NeoTreeSymbolicLinkTarget = { fg = c.blue },

    -- Notify
    NotifyERRORBorder = { fg = c.error },
    NotifyWARNBorder = { fg = c.warning },
    NotifyINFOBorder = { fg = c.info },
    NotifyDEBUGBorder = { fg = c.comment },
    NotifyTRACEBorder = { fg = c.comment },
    NotifyERRORIcon = { fg = c.error },
    NotifyWARNIcon = { fg = c.warning },
    NotifyINFOIcon = { fg = c.info },
    NotifyDEBUGIcon = { fg = c.comment },
    NotifyTRACEIcon = { fg = c.comment },
    NotifyERRORTitle = { fg = c.error },
    NotifyWARNTitle = { fg = c.warning },
    NotifyINFOTitle = { fg = c.info },
    NotifyDEBUGTitle = { fg = c.comment },
    NotifyTRACETitle = { fg = c.comment },

    -- Trouble
    TroubleNormal = { fg = c.fg, bg = c.bg_dark },
    TroubleText = { fg = c.fg },
    TroubleCount = { fg = c.green, bg = c.bg_highlight },
    TroubleError = { fg = c.error },
    TroubleWarning = { fg = c.warning },
    TroubleInformation = { fg = c.info },
    TroubleHint = { fg = c.hint },

    -- Flash
    FlashLabel = { fg = c.bg, bg = c.orange, bold = true },
    FlashBackdrop = { fg = c.comment },
    FlashCurrent = { fg = c.bg, bg = c.green },
    FlashMatch = { fg = c.bg, bg = c.cyan },

    -- Todo-comments
    TodoBgTODO = { fg = c.bg, bg = c.blue, bold = true },
    TodoBgFIX = { fg = c.bg, bg = c.red, bold = true },
    TodoBgHACK = { fg = c.bg, bg = c.orange, bold = true },
    TodoBgWARN = { fg = c.bg, bg = c.orange, bold = true },
    TodoBgNOTE = { fg = c.bg, bg = c.green, bold = true },
    TodoBgTEST = { fg = c.bg, bg = c.cyan, bold = true },
    TodoFgTODO = { fg = c.blue },
    TodoFgFIX = { fg = c.red },
    TodoFgHACK = { fg = c.orange },
    TodoFgWARN = { fg = c.orange },
    TodoFgNOTE = { fg = c.green },
    TodoFgTEST = { fg = c.cyan },
    TodoSignTODO = { fg = c.blue },
    TodoSignFIX = { fg = c.red },
    TodoSignHACK = { fg = c.orange },
    TodoSignWARN = { fg = c.orange },
    TodoSignNOTE = { fg = c.green },
    TodoSignTEST = { fg = c.cyan },

    -- Mini
    MiniCursorword = { bg = c.bg_highlight },
    MiniCursorwordCurrent = { bg = c.bg_highlight },
    MiniIndentscopeSymbol = { fg = c.blue_dark },
    MiniIndentscopePrefix = { nocombine = true },
    MiniJump = { fg = c.bg, bg = c.orange },
    MiniJump2dSpot = { fg = c.orange, bold = true, nocombine = true },
    MiniStatuslineDevinfo = { fg = c.fg, bg = c.bg_highlight },
    MiniStatuslineFileinfo = { fg = c.fg, bg = c.bg_highlight },
    MiniStatuslineFilename = { fg = c.fg, bg = c.bg_dark },
    MiniStatuslineInactive = { fg = c.comment, bg = c.bg_dark },
    MiniStatuslineModeCommand = { fg = c.bg, bg = c.orange, bold = true },
    MiniStatuslineModeInsert = { fg = c.bg, bg = c.green, bold = true },
    MiniStatuslineModeNormal = { fg = c.bg, bg = c.blue, bold = true },
    MiniStatuslineModeOther = { fg = c.bg, bg = c.cyan, bold = true },
    MiniStatuslineModeReplace = { fg = c.bg, bg = c.red, bold = true },
    MiniStatuslineModeVisual = { fg = c.bg, bg = c.green_bright, bold = true },
    MiniSurround = { fg = c.bg, bg = c.orange },
    MiniTablineCurrent = { fg = c.fg, bg = c.bg },
    MiniTablineFill = { bg = c.bg_dark },
    MiniTablineHidden = { fg = c.comment, bg = c.bg_dark },
    MiniTablineModifiedCurrent = { fg = c.orange, bg = c.bg },
    MiniTablineModifiedHidden = { fg = c.orange, bg = c.bg_dark },
    MiniTablineModifiedVisible = { fg = c.orange, bg = c.bg },
    MiniTablineTabpagesection = { fg = c.green, bg = c.bg_dark },
    MiniTablineVisible = { fg = c.fg, bg = c.bg },

    -- render-markdown: headings in green for emphasis and blue for structure,
    -- each underlined in grey; the text colour carries the hierarchy
    RenderMarkdownH1 = { fg = c.green, bold = true },
    RenderMarkdownH2 = { fg = c.green, bold = true },
    RenderMarkdownH3 = { fg = c.cyan },
    RenderMarkdownH4 = { fg = c.blue },
    RenderMarkdownH5 = { fg = c.blue },
    RenderMarkdownH6 = { fg = c.comment },
    RenderMarkdownH1Bg = { sp = c.comment, underline = true },
    RenderMarkdownH2Bg = { sp = c.comment, underline = true },
    RenderMarkdownH3Bg = { sp = c.comment, underline = true },
    RenderMarkdownH4Bg = { sp = c.comment, underline = true },
    RenderMarkdownH5Bg = { sp = c.comment, underline = true },
    RenderMarkdownH6Bg = { sp = c.comment, underline = true },
  }
end

return M
