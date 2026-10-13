-- Configure neo-tree: window picker + jj source
return {
  -- Add window picker plugin
  {
    "s1n7ax/nvim-window-picker",
    version = "2.*",
    config = function()
      require("window-picker").setup({
        hint = "floating-big-letter",
        filter_rules = {
          include_current_win = false,
          autoselect_one = true,
          bo = {
            filetype = { "neo-tree", "neo-tree-popup", "notify" },
            buftype = { "terminal", "quickfix" },
          },
        },
      })
    end,
  },

  -- neo-tree-jj: jj status source for neo-tree.
  -- FORK (lkingland/neo-tree-jj.nvim): upstream Cretezy is dead (last commit 2024)
  -- and get_changes()'s diff revset is hardcoded, so we carry a one-commit patch
  -- widening it to `--from "roots(ancestors(@, 2))"` — matches vcsigns'
  -- target_commit=1 (diff from grandparent, persists across `jj new`).
  -- To pull upstream later: `git fetch upstream && git rebase upstream/main` in the
  -- plugin dir, then `git push --force-with-lease origin main`.
  {
    "lkingland/neo-tree-jj.nvim",
    dependencies = { "nvim-neo-tree/neo-tree.nvim" },
  },

  -- Configure neo-tree
  {
    "nvim-neo-tree/neo-tree.nvim",
    keys = {
      -- Override LazyVim defaults: <leader>e = focus, <leader>E = toggle
      {
        "<leader>e",
        function()
          require("neo-tree.command").execute({ action = "focus", dir = LazyVim.root() })
        end,
        desc = "Focus Explorer (Root Dir)",
      },
      {
        "<leader>E",
        function()
          require("neo-tree.command").execute({ toggle = true, dir = LazyVim.root() })
        end,
        desc = "Toggle Explorer (Root Dir)",
      },
      {
        "<leader>ej",
        "<cmd>Neotree jj<cr>",
        desc = "Explorer JJ Status",
      },
    },
    opts = function(_, opts)
      -- Register the jj source
      opts.sources = opts.sources or {}
      if not vim.tbl_contains(opts.sources, "jj") then
        table.insert(opts.sources, "jj")
      end

      -- Replace git_status tab with jj tab in source selector (if in a jj repo)
      local ok, jj_utils = pcall(require, "neo-tree.sources.jj.utils")
      if ok and jj_utils.get_repository_root() then
        if opts.source_selector and opts.source_selector.sources then
          for i, source in ipairs(opts.source_selector.sources) do
            if source.source == "git_status" then
              table.remove(opts.source_selector.sources, i)
              break
            end
          end
          table.insert(opts.source_selector.sources, {
            display_name = "  JJ",
            source = "jj",
          })
        end
      end

      -- Disable git_status GLOBALLY. This must be the global flag, not
      -- `filesystem.enable_git_status` — neo-tree does `config.enable_git_status =
      -- config.enable_git_status or global_config.enable_git_status`, so an explicit
      -- `false` is flipped back to the global `true` (and the fetch reads global anyway).
      -- Why off: these dirs are *secondary* jj workspaces, invisible to the colocated
      -- parent ~/hamr/.git except as "untracked" — so git_status paints every node `?`.
      -- Real jj status is supplied by jj instead — see the jj_status wiring below.
      opts.enable_git_status = false

      -- Show jj change status (A/M/D) on the FILESYSTEM tree. git_status is git-only
      -- and wrong in jj workspaces (disabled above), so we repurpose that now-inert
      -- render slot: override the filesystem `git_status` component with neo-tree-jj's
      -- jj-backed renderer, fed by a per-workspace lookup computed before each render
      -- (changed files rolled up to ancestor dirs so folders flag changes too). The
      -- lookup is cached per jj-root and invalidated on save, so renders stay snappy.
      if ok then
        local jj_root_cache = {} -- state.path -> jj_root | false
        local jj_changes_cache = {} -- jj_root -> { [abs_path] = "A"|"M"|"D" } (+ dirs)

        local function jj_root_of(path)
          if jj_root_cache[path] == nil then
            jj_root_cache[path] = jj_utils.get_repository_root(path) or false
          end
          return jj_root_cache[path] or nil
        end

        local function jj_lookup(root)
          if jj_changes_cache[root] then
            return jj_changes_cache[root]
          end
          local changes = jj_utils.get_changes(root) or {}
          local agg = {}
          for path, status in pairs(changes) do
            agg[path] = status
          end
          for path in pairs(changes) do -- roll changes up to ancestor directories
            local dir = vim.fn.fnamemodify(path, ":h")
            while #dir > #root do
              if agg[dir] then
                break
              end
              agg[dir] = "M" -- "this folder contains changes"
              dir = vim.fn.fnamemodify(dir, ":h")
            end
          end
          jj_changes_cache[root] = agg
          return agg
        end

        opts.filesystem = opts.filesystem or {}
        opts.filesystem.components = opts.filesystem.components or {}
        opts.filesystem.components.git_status = function(config, node, state)
          return jj_utils.get_jj_status(config, node, state)
        end

        opts.event_handlers = opts.event_handlers or {}
        table.insert(opts.event_handlers, {
          event = "before_render",
          handler = function(state)
            if state.name ~= "filesystem" then
              return
            end
            local root = jj_root_of(state.path)
            state.jj_status_lookup = root and jj_lookup(root) or nil
          end,
        })

        -- jj status icons (reusing the git_status component's symbol slots)
        opts.default_component_configs = opts.default_component_configs or {}
        opts.default_component_configs.git_status = vim.tbl_deep_extend(
          "force",
          opts.default_component_configs.git_status or {},
          {
            symbols = {
              added = "✚",
              modified = "●",
              deleted = "✖",
              renamed = "",
              staged = "",
              unstaged = "",
              untracked = "",
              ignored = "",
              conflict = "",
            },
          }
        )

        -- invalidate caches + refresh the tree when the working copy changes
        vim.api.nvim_create_autocmd({ "BufWritePost", "FocusGained" }, {
          group = vim.api.nvim_create_augroup("NeoTreeJjStatus", { clear = true }),
          callback = function()
            jj_root_cache = {}
            jj_changes_cache = {}
            pcall(function()
              require("neo-tree.sources.manager").refresh("filesystem")
            end)
          end,
        })
      end

      -- Merge remaining opts
      opts.open_files_do_not_replace_types = { "terminal", "Trouble", "trouble", "qf", "Outline" }
      opts.window = vim.tbl_deep_extend("force", opts.window or {}, {
        mappings = {
          ["<cr>"] = "open_with_window_picker",
          ["o"] = "open_with_window_picker",
          ["s"] = "split_with_window_picker",
          ["v"] = "vsplit_with_window_picker",
          ["/"] = "none",
        },
      })
    end,
  },
}
