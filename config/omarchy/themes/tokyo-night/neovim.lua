-- Onsen overlay on Tokyo Night: the stock tokyonight-night colorscheme, made
-- transparent so Neovim shows the terminal's black, sidebars and floats too.
return {
	{
		"folke/tokyonight.nvim",
		priority = 1000,
		opts = {
			style = "night",
			transparent = true,
			styles = { sidebars = "transparent", floats = "transparent" },
		},
	},
	{
		"LazyVim/LazyVim",
		opts = {
			colorscheme = "tokyonight-night",
		},
	},
}
