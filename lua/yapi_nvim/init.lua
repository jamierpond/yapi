-- lua/yapi_nvim/init.lua
--
-- yapi Neovim integration:
-- - :YapiWatch opens pretty TUI watch mode in a terminal
-- - :YapiRun runs a single request in a split
-- - LSP support for completions and diagnostics

local M = {}

local term_buf = nil
local term_win = nil

local function open_term_window()
  -- Check if window still exists and is valid
  if term_win and vim.api.nvim_win_is_valid(term_win) then
    vim.api.nvim_set_current_win(term_win)
    return term_win
  end

  -- Create a vertical split on the right
  vim.cmd("rightbelow vsplit")
  term_win = vim.api.nvim_get_current_win()

  -- Window options
  vim.wo[term_win].number = false
  vim.wo[term_win].relativenumber = false
  vim.wo[term_win].signcolumn = "no"
  vim.wo[term_win].foldcolumn = "0"

  return term_win
end

local function close_term()
  if term_win and vim.api.nvim_win_is_valid(term_win) then
    vim.api.nvim_win_close(term_win, true)
  end
  term_win = nil
  term_buf = nil
end

local function start_watch(filepath)
  if filepath == "" then
    vim.notify("[yapi] Buffer has no file name", vim.log.levels.ERROR)
    return
  end

  if not filepath:match("%.yapi$") and
     not filepath:match("%.yapi%.yml$") and
     not filepath:match("%.yapi%.yaml$") and
     not filepath:match("yapi%.config%.yml$") and
     not filepath:match("yapi%.config%.yaml$")
  then
    vim.notify("[yapi] Not a yapi config file", vim.log.levels.WARN)
    return
  end

  -- Save if modified
  if vim.bo.modified then
    vim.cmd("write")
  end

  -- Close existing term if any
  close_term()

  -- Open window and launch terminal with yapi watch --pretty
  local win = open_term_window()
  term_buf = vim.api.nvim_create_buf(false, true)
  vim.api.nvim_win_set_buf(win, term_buf)

  -- Start terminal with yapi watch (pretty mode if configured)
  local cmd = { "yapi", "watch", filepath }
  if M._opts.pretty then
    table.insert(cmd, "--pretty")
  end
  vim.fn.termopen(cmd, {
    on_exit = function()
      vim.schedule(function()
        close_term()
      end)
    end,
  })

  -- Stay in normal mode, don't enter insert mode in terminal
  vim.cmd("stopinsert")
end

local function run_once(filepath)
  if filepath == "" then
    vim.notify("[yapi] Buffer has no file name", vim.log.levels.ERROR)
    return
  end

  if not filepath:match("%.yapi$") and
     not filepath:match("%.yapi%.yml$") and
     not filepath:match("%.yapi%.yaml$") and
     not filepath:match("yapi%.config%.yml$") and
     not filepath:match("yapi%.config%.yaml$")
  then
    vim.notify("[yapi] Not a yapi config file", vim.log.levels.WARN)
    return
  end

  if vim.bo.modified then
    vim.cmd("write")
  end

  -- Close existing term if any
  close_term()

  -- Open window and run yapi
  local win = open_term_window()
  term_buf = vim.api.nvim_create_buf(false, true)
  vim.api.nvim_win_set_buf(win, term_buf)

  vim.fn.termopen({ "yapi", "run", filepath }, {
    on_exit = function()
      -- Keep the buffer open to show results
    end,
  })

  vim.cmd("stopinsert")
end

function M.watch()
  start_watch(vim.api.nvim_buf_get_name(0))
end

function M.run()
  run_once(vim.api.nvim_buf_get_name(0))
end

function M.stop()
  close_term()
end

function M.toggle()
  if term_win and vim.api.nvim_win_is_valid(term_win) then
    close_term()
  else
    start_watch(vim.api.nvim_buf_get_name(0))
  end
end

function M.setup(opts)
  opts = opts or {}
  M._opts = {
    lsp = opts.lsp ~= false,
    pretty = opts.pretty == true,  -- default false
    watch_on_save = opts.watch_on_save == true,  -- default false
  }

  -- Commands
  vim.api.nvim_create_user_command("YapiWatch", function()
    M.watch()
  end, { desc = "Start yapi watch mode in terminal" })

  vim.api.nvim_create_user_command("YapiRun", function()
    M.run()
  end, { desc = "Run yapi once in terminal" })

  vim.api.nvim_create_user_command("YapiStop", function()
    M.stop()
  end, { desc = "Close yapi terminal" })

  -- Auto-start watch on save if configured
  if M._opts.watch_on_save then
    vim.api.nvim_create_autocmd("BufWritePost", {
      pattern = { "*.yapi.yml", "*.yapi.yaml", "*.yapi", "yapi.config.yml", "yapi.config.yaml" },
      callback = function()
        M.watch()
      end,
      desc = "Start yapi watch on save",
    })
  end

  -- Setup LSP for yapi files
  if M._opts.lsp then
    vim.lsp.config.yapi = {
      cmd = { "yapi", "lsp" },
      filetypes = { "yaml.yapi" },
      root_markers = { "yapi.config.yml", "yapi.config.yaml", ".git" },
    }
    vim.lsp.enable("yapi")

    -- Set filetype to yaml.yapi for yapi config files
    vim.api.nvim_create_autocmd({ "BufReadPost", "BufNewFile" }, {
      pattern = {
        "*.yapi.yml",
        "*.yapi.yaml",
        "*.yapi",
        "yapi.config.yml",
        "yapi.config.yaml"
      },
      callback = function()
        vim.bo.filetype = "yaml.yapi"
      end,
      desc = "Set filetype for yapi config files",
    })
  end

  -- Clean up on vim exit
  vim.api.nvim_create_autocmd("VimLeavePre", {
    callback = function()
      close_term()
    end,
  })
end

return M
