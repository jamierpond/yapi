-- lua/yapi_nvim/init.lua
--
-- yapi Neovim integration:
-- - :YapiWatch starts watch mode in a side panel
-- - :YapiRun runs a single request
-- - :YapiStop stops watch mode
-- - LSP support for completions and diagnostics

local M = {}

local RESULT_BUF_NAME = "yapi://result"
local watch_job_id = nil
local watch_filepath = nil

local function get_result_buf()
  for _, buf in ipairs(vim.api.nvim_list_bufs()) do
    if vim.api.nvim_buf_is_loaded(buf) then
      local name = vim.api.nvim_buf_get_name(buf)
      if name:match(RESULT_BUF_NAME .. "$") then
        return buf
      end
    end
  end

  local buf = vim.api.nvim_create_buf(false, true)
  vim.api.nvim_buf_set_name(buf, RESULT_BUF_NAME)
  vim.bo[buf].bufhidden = "hide"
  vim.bo[buf].buftype = "nofile"
  vim.bo[buf].swapfile = false
  vim.bo[buf].filetype = "yapiresult"
  return buf
end

local function open_result_window()
  local buf = get_result_buf()

  for _, win in ipairs(vim.api.nvim_list_wins()) do
    if vim.api.nvim_win_get_buf(win) == buf then
      return win, buf
    end
  end

  vim.cmd("rightbelow vsplit")
  local win = vim.api.nvim_get_current_win()
  vim.api.nvim_win_set_buf(win, buf)
  vim.wo[win].wrap = false
  vim.wo[win].number = false
  vim.wo[win].relativenumber = false
  vim.wo[win].signcolumn = "no"
  vim.wo[win].foldcolumn = "0"
  vim.wo[win].statusline = "%#Comment# yapi %*"
  return win, buf
end

local function close_result_window()
  local buf = nil
  for _, b in ipairs(vim.api.nvim_list_bufs()) do
    if vim.api.nvim_buf_is_loaded(b) then
      local name = vim.api.nvim_buf_get_name(b)
      if name:match(RESULT_BUF_NAME .. "$") then
        buf = b
        break
      end
    end
  end

  if buf then
    for _, win in ipairs(vim.api.nvim_list_wins()) do
      if vim.api.nvim_win_get_buf(win) == buf then
        vim.api.nvim_win_close(win, true)
        return
      end
    end
  end
end

local function stop_watch()
  if watch_job_id then
    vim.fn.jobstop(watch_job_id)
    watch_job_id = nil
    watch_filepath = nil
  end
end

local function strip_ansi(str)
  if not str then return "" end
  return str:gsub("\027%[[%d;]*m", "")
end

local function start_watch(filepath)
  if filepath == "" then
    vim.notify("[yapi] Buffer has no file name", vim.log.levels.ERROR)
    return
  end

  if not filepath:match("%.yapi$") and
     not filepath:match("%.yapi%.yml$") and
     not filepath:match("%.yapi%.yaml$")
  then
    vim.notify("[yapi] Not a yapi config file", vim.log.levels.WARN)
    return
  end

  -- Stop existing watch if any
  stop_watch()

  -- Save if modified
  if vim.bo.modified then
    vim.cmd("write")
  end

  watch_filepath = filepath
  local _, buf = open_result_window()
  vim.bo[buf].modifiable = true
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, { "Starting watch..." })

  local output_lines = {}

  watch_job_id = vim.fn.jobstart({ "yapi", "watch", filepath }, {
    stdout_buffered = false,
    stderr_buffered = false,

    on_stdout = function(_, data)
      if not data then return end
      vim.schedule(function()
        if not vim.api.nvim_buf_is_valid(buf) then return end

        for _, line in ipairs(data) do
          -- Clear screen escape sequence detection
          if line:match("\027%[H\027%[2J") or line:match("\027%[2J") then
            output_lines = {}
          else
            local clean = strip_ansi(line)
            if clean ~= "" or #output_lines > 0 then
              table.insert(output_lines, clean)
            end
          end
        end

        vim.bo[buf].modifiable = true
        vim.api.nvim_buf_set_lines(buf, 0, -1, false, output_lines)
        vim.bo[buf].modifiable = false
      end)
    end,

    on_stderr = function(_, data)
      if not data then return end
      vim.schedule(function()
        if not vim.api.nvim_buf_is_valid(buf) then return end
        for _, line in ipairs(data) do
          local clean = strip_ansi(line)
          if clean ~= "" then
            table.insert(output_lines, clean)
          end
        end
        vim.bo[buf].modifiable = true
        vim.api.nvim_buf_set_lines(buf, 0, -1, false, output_lines)
        vim.bo[buf].modifiable = false
      end)
    end,

    on_exit = function(_, code)
      vim.schedule(function()
        watch_job_id = nil
        watch_filepath = nil
        if code ~= 0 and code ~= 143 then -- 143 = SIGTERM (normal stop)
          vim.notify("[yapi] watch exited with code " .. code, vim.log.levels.WARN)
        end
      end)
    end,
  })

  if watch_job_id <= 0 then
    vim.notify("[yapi] Failed to start watch", vim.log.levels.ERROR)
    watch_job_id = nil
  end
end

local function run_once(filepath)
  if filepath == "" then
    vim.notify("[yapi] Buffer has no file name", vim.log.levels.ERROR)
    return
  end

  if not filepath:match("%.yapi$") and
     not filepath:match("%.yapi%.yml$") and
     not filepath:match("%.yapi%.yaml$")
  then
    vim.notify("[yapi] Not a yapi config file", vim.log.levels.WARN)
    return
  end

  if vim.bo.modified then
    vim.cmd("write")
  end

  local _, buf = open_result_window()
  vim.bo[buf].modifiable = true
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, { "Running..." })

  vim.fn.jobstart({ "yapi", "run", filepath }, {
    stdout_buffered = true,
    stderr_buffered = true,

    on_stdout = function(_, data)
      if not data then return end
      vim.schedule(function()
        if not vim.api.nvim_buf_is_valid(buf) then return end
        local lines = {}
        for _, line in ipairs(data) do
          table.insert(lines, strip_ansi(line))
        end
        vim.bo[buf].modifiable = true
        vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
        vim.bo[buf].modifiable = false
      end)
    end,

    on_stderr = function(_, data)
      if not data then return end
      vim.schedule(function()
        if not vim.api.nvim_buf_is_valid(buf) then return end
        vim.bo[buf].modifiable = true
        local existing = vim.api.nvim_buf_get_lines(buf, 0, -1, false)
        for _, line in ipairs(data) do
          local clean = strip_ansi(line)
          if clean ~= "" then
            table.insert(existing, clean)
          end
        end
        vim.api.nvim_buf_set_lines(buf, 0, -1, false, existing)
        vim.bo[buf].modifiable = false
      end)
    end,

    on_exit = function(_, code)
      vim.schedule(function()
        if code ~= 0 then
          vim.notify("[yapi] exited with code " .. code, vim.log.levels.ERROR)
        end
      end)
    end,
  })
end

function M.watch()
  start_watch(vim.api.nvim_buf_get_name(0))
end

function M.run()
  run_once(vim.api.nvim_buf_get_name(0))
end

function M.stop()
  stop_watch()
  close_result_window()
end

function M.is_watching()
  return watch_job_id ~= nil
end

function M.setup(opts)
  opts = opts or {}
  local enable_lsp = opts.lsp ~= false

  -- Commands
  vim.api.nvim_create_user_command("YapiWatch", function()
    M.watch()
  end, { desc = "Start yapi watch mode for current file" })

  vim.api.nvim_create_user_command("YapiRun", function()
    M.run()
  end, { desc = "Run yapi once for current file" })

  vim.api.nvim_create_user_command("YapiStop", function()
    M.stop()
  end, { desc = "Stop yapi watch mode" })

  -- Setup LSP for yapi files
  if enable_lsp then
    vim.lsp.config.yapi = {
      cmd = { "yapi", "lsp" },
      filetypes = { "yaml.yapi" },
      root_markers = { ".git" },
    }
    vim.lsp.enable("yapi")

    -- Set filetype to yaml.yapi for yapi config files
    vim.api.nvim_create_autocmd({ "BufReadPost", "BufNewFile" }, {
      pattern = { "*.yapi.yml", "*.yapi.yaml" },
      callback = function()
        vim.bo.filetype = "yaml.yapi"
      end,
      desc = "Set filetype for yapi config files",
    })
  end

  -- Clean up watch on vim exit
  vim.api.nvim_create_autocmd("VimLeavePre", {
    callback = function()
      stop_watch()
    end,
  })
end

return M

