-- Onsen treesitter highlight groups
-- Semantic syntax highlighting via nvim-treesitter

local M = {}

function M.get(c)
  return {
    -- Misc
    ["@comment"] = { fg = c.comment },
    ["@comment.documentation"] = { fg = c.comment },
    ["@error"] = { fg = c.red },
    ["@none"] = {},
    ["@preproc"] = { fg = c.blue },
    ["@define"] = { fg = c.blue },

    -- Punctuation (dimmed to reduce noise)
    ["@punctuation.delimiter"] = { fg = c.grey },
    ["@punctuation.bracket"] = { fg = c.grey },
    ["@punctuation.special"] = { fg = c.grey },

    -- Literals
    ["@string"] = { fg = c.cyan },
    ["@string.documentation"] = { fg = c.cyan },
    ["@string.regex"] = { fg = c.fg },
    ["@string.escape"] = { fg = c.cyan_bright },
    ["@string.special"] = { fg = c.cyan },
    ["@string.special.symbol"] = { fg = c.cyan_bright },
    ["@string.special.url"] = { fg = c.blue, underline = true },
    ["@string.special.path"] = { fg = c.cyan },

    ["@character"] = { fg = c.cyan },
    ["@character.special"] = { fg = c.cyan },

    ["@number"] = { fg = c.fg },
    ["@number.float"] = { fg = c.fg },

    ["@boolean"] = { fg = c.fg },

    -- Functions (neutral: function names are just identifiers)
    ["@function"] = { fg = c.fg },
    ["@function.builtin"] = { fg = c.fg },
    ["@function.call"] = { fg = c.fg },
    ["@function.macro"] = { fg = c.fg },
    ["@function.method"] = { fg = "#66E8A0" },
    ["@function.method.call"] = { fg = "#66E8A0" },

    -- Keywords (blue: control flow + reserved words)
    ["@keyword"] = { fg = c.blue },
    ["@keyword.conditional"] = { fg = c.blue },
    ["@keyword.conditional.ternary"] = { fg = c.blue },
    ["@keyword.coroutine"] = { fg = c.blue },
    ["@keyword.debug"] = { fg = c.orange },
    ["@keyword.directive"] = { fg = c.blue },
    ["@keyword.directive.define"] = { fg = c.blue },
    ["@keyword.exception"] = { fg = c.blue },
    ["@keyword.function"] = { fg = c.blue },
    ["@keyword.import"] = { fg = c.blue },
    ["@keyword.modifier"] = { fg = c.blue },
    ["@keyword.operator"] = { fg = c.blue },
    ["@keyword.repeat"] = { fg = c.blue },
    ["@keyword.return"] = { fg = c.blue },
    ["@keyword.storage"] = { fg = c.blue },
    ["@keyword.type"] = { fg = c.blue },

    -- Identifiers
    ["@variable"] = { fg = c.fg },
    ["@variable.builtin"] = { fg = c.fg },
    ["@variable.member"] = { fg = "#66E8A0" },
    ["@variable.parameter"] = { fg = c.fg },
    ["@variable.parameter.builtin"] = { fg = c.fg },

    ["@constant"] = { fg = c.fg },
    ["@constant.builtin"] = { fg = c.fg },
    ["@constant.macro"] = { fg = c.fg },

    ["@module"] = { fg = c.fg },
    ["@module.builtin"] = { fg = c.fg },

    ["@label"] = { fg = c.blue },

    -- Types (italic only, no color accent)
    ["@type"] = { fg = c.fg, italic = true },
    ["@type.builtin"] = { fg = c.fg, italic = true },
    ["@type.definition"] = { fg = c.fg, italic = true },
    ["@type.qualifier"] = { fg = c.blue },

    ["@attribute"] = { fg = c.orange },
    ["@attribute.builtin"] = { fg = c.orange },

    ["@property"] = { fg = "#66E8A0" },

    -- Operators
    ["@operator"] = { fg = c.fg_dark },

    -- Tags (HTML, JSX, etc.)
    ["@tag"] = { fg = c.green },
    ["@tag.builtin"] = { fg = c.green },
    ["@tag.attribute"] = { fg = c.fg },
    ["@tag.delimiter"] = { fg = c.fg_dark },

    -- Constructor (type-related)
    ["@constructor"] = { fg = c.fg, italic = true },

    -- Special
    ["@annotation"] = { fg = c.orange },

    -- Markup (Markdown, etc.)
    ["@markup"] = { fg = c.fg },
    ["@markup.heading"] = { fg = c.green, bold = true },
    ["@markup.heading.1"] = { fg = c.green, bold = true },
    ["@markup.heading.2"] = { fg = c.green, bold = true },
    ["@markup.heading.3"] = { fg = c.green, bold = true },
    ["@markup.heading.4"] = { fg = c.green },
    ["@markup.heading.5"] = { fg = c.green },
    ["@markup.heading.6"] = { fg = c.green },
    ["@markup.italic"] = { italic = true },
    ["@markup.strong"] = { bold = true },
    ["@markup.strikethrough"] = { strikethrough = true },
    ["@markup.underline"] = { underline = true },
    ["@markup.quote"] = { fg = c.comment, italic = true },
    ["@markup.math"] = { fg = c.cyan },
    ["@markup.environment"] = { fg = c.orange },
    ["@markup.link"] = { fg = c.blue },
    ["@markup.link.label"] = { fg = c.blue },
    ["@markup.link.url"] = { fg = c.blue, underline = true },
    ["@markup.raw"] = { fg = c.orange },
    ["@markup.raw.block"] = { fg = c.fg },
    ["@markup.list"] = { fg = c.fg_dark },
    ["@markup.list.checked"] = { fg = c.green },
    ["@markup.list.unchecked"] = { fg = c.comment },

    -- Diff
    ["@diff.plus"] = { fg = c.git_add },
    ["@diff.minus"] = { fg = c.git_delete },
    ["@diff.delta"] = { fg = c.git_change },
  }
end

return M
