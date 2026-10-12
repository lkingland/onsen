-- Onsen LSP highlight groups
-- Diagnostics, references, and semantic tokens

local M = {}

function M.get(c)
  return {
    -- Diagnostics
    DiagnosticError = { fg = c.error },
    DiagnosticWarn = { fg = c.warning },
    DiagnosticInfo = { fg = c.info },
    DiagnosticHint = { fg = c.hint },
    DiagnosticOk = { fg = c.green },

    DiagnosticVirtualTextError = { fg = c.error, bg = c.bg_highlight },
    DiagnosticVirtualTextWarn = { fg = c.warning, bg = c.bg_highlight },
    DiagnosticVirtualTextInfo = { fg = c.info, bg = c.bg_highlight },
    DiagnosticVirtualTextHint = { fg = c.hint, bg = c.bg_highlight },
    DiagnosticVirtualTextOk = { fg = c.green, bg = c.bg_highlight },

    DiagnosticUnderlineError = { undercurl = true, sp = c.error },
    DiagnosticUnderlineWarn = { undercurl = true, sp = c.warning },
    DiagnosticUnderlineInfo = { undercurl = true, sp = c.info },
    DiagnosticUnderlineHint = { undercurl = true, sp = c.hint },
    DiagnosticUnderlineOk = { undercurl = true, sp = c.green },

    DiagnosticFloatingError = { fg = c.error },
    DiagnosticFloatingWarn = { fg = c.warning },
    DiagnosticFloatingInfo = { fg = c.info },
    DiagnosticFloatingHint = { fg = c.hint },
    DiagnosticFloatingOk = { fg = c.green },

    DiagnosticSignError = { fg = c.error },
    DiagnosticSignWarn = { fg = c.warning },
    DiagnosticSignInfo = { fg = c.info },
    DiagnosticSignHint = { fg = c.hint },
    DiagnosticSignOk = { fg = c.green },

    -- LSP References (slight lift above bg_highlight for visibility)
    LspReferenceText = { bg = "#152A20" },
    LspReferenceRead = { bg = "#152A20" },
    LspReferenceWrite = { bg = c.bg_visual },

    -- LSP Signature
    LspSignatureActiveParameter = { bg = c.bg_visual, bold = true },

    -- LSP CodeLens
    LspCodeLens = { fg = c.comment },
    LspCodeLensSeparator = { fg = c.comment },

    -- LSP Inlay Hints
    LspInlayHint = { fg = c.comment, bg = c.bg_highlight },

    -- LSP Info window
    LspInfoBorder = { fg = c.green },

    -- Semantic Tokens
    ["@lsp.type.boolean"] = { link = "@boolean" },
    ["@lsp.type.builtinType"] = { link = "@type.builtin" },
    ["@lsp.type.comment"] = { link = "@comment" },
    ["@lsp.type.decorator"] = { link = "@attribute" },
    ["@lsp.type.deriveHelper"] = { link = "@attribute" },
    ["@lsp.type.enum"] = { link = "@type" },
    ["@lsp.type.enumMember"] = { link = "@constant" },
    ["@lsp.type.escapeSequence"] = { link = "@string.escape" },
    ["@lsp.type.formatSpecifier"] = { link = "@markup.list" },
    ["@lsp.type.function"] = { link = "@function" },
    ["@lsp.type.function.go"] = {},  -- Let treesitter handle Go (distinguishes pkg.Func from func)
    ["@lsp.type.generic"] = { link = "@variable" },
    ["@lsp.type.interface"] = { link = "@type" },
    ["@lsp.type.keyword"] = { link = "@keyword" },
    ["@lsp.type.lifetime"] = { link = "@keyword.storage" },
    ["@lsp.type.macro"] = { link = "@function.macro" },
    ["@lsp.type.method"] = { link = "@function.method" },
    ["@lsp.type.namespace"] = { link = "@module" },
    ["@lsp.type.namespace.go"] = {},  -- Let treesitter @string handle Go imports
    ["@lsp.type.number"] = { link = "@number" },
    ["@lsp.type.operator"] = { link = "@operator" },
    ["@lsp.type.parameter"] = { link = "@variable.parameter" },
    ["@lsp.type.property"] = { link = "@property" },
    ["@lsp.type.field"] = { link = "@property" },
    ["@lsp.type.member"] = { link = "@variable.member" },
    ["@lsp.type.selfKeyword"] = { link = "@variable.builtin" },
    ["@lsp.type.selfTypeKeyword"] = { link = "@variable.builtin" },
    ["@lsp.type.string"] = { link = "@string" },
    ["@lsp.type.struct"] = { link = "@type" },
    ["@lsp.type.type"] = { link = "@type" },
    ["@lsp.type.typeAlias"] = { link = "@type.definition" },
    ["@lsp.type.typeParameter"] = { link = "@type" },
    ["@lsp.type.unresolvedReference"] = { undercurl = true, sp = c.error },
    ["@lsp.type.variable"] = {},  -- Use treesitter

    -- Semantic Token Modifiers
    ["@lsp.typemod.class.defaultLibrary"] = { link = "@type.builtin" },
    ["@lsp.typemod.enum.defaultLibrary"] = { link = "@type.builtin" },
    ["@lsp.typemod.enumMember.defaultLibrary"] = { link = "@constant.builtin" },
    ["@lsp.typemod.function.defaultLibrary"] = { link = "@function.builtin" },
    ["@lsp.typemod.keyword.async"] = { link = "@keyword.coroutine" },
    ["@lsp.typemod.keyword.injected"] = { link = "@keyword" },
    ["@lsp.typemod.macro.defaultLibrary"] = { link = "@function.builtin" },
    ["@lsp.typemod.method.defaultLibrary"] = { link = "@function.builtin" },
    ["@lsp.typemod.operator.injected"] = { link = "@operator" },
    ["@lsp.typemod.string.injected"] = { link = "@string" },
    ["@lsp.typemod.struct.defaultLibrary"] = { link = "@type.builtin" },
    ["@lsp.typemod.type.defaultLibrary"] = { link = "@type.builtin" },
    ["@lsp.typemod.typeAlias.defaultLibrary"] = { link = "@type.builtin" },
    ["@lsp.typemod.variable.callable"] = { link = "@function" },
    ["@lsp.typemod.variable.defaultLibrary"] = { link = "@variable.builtin" },
    ["@lsp.typemod.variable.injected"] = { link = "@variable" },
    ["@lsp.typemod.variable.static"] = { link = "@constant" },
    ["@lsp.typemod.variable.globalScope"] = {},
    ["@lsp.typemod.variable.readonly"] = { link = "@variable" },  -- Constants use regular green
    ["@lsp.typemod.variable.readonly.go"] = { link = "@variable" },
  }
end

return M
