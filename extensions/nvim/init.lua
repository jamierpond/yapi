-- lua/yapi_nvim/init.lua
--
-- yapi Neovim integration:
-- - :YapiRun runs `yapi -c <current file>`
-- - shows output in a right-hand split
-- - optional hot reload on save

local M = {}

local RESULT_BUF_NAME = "yapi://result"

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
  vim.api.nvim_buf_set_option(buf, "bufhidden", "wipe")
  vim.api.nvim_buf_set_option(buf, "filetype", "yapiresult")
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
  vim.api.nvim_win_set_option(win, "wrap", false)
  vim.api.nvim_win_set_option(win, "number", false)
  vim.api.nvim_win_set_option(win, "relativenumber", false)
  return win, buf
end

local function run_yapi_for_current()
  local filepath = vim.api.nvim_buf_get_name(0)
  if filepath == "" then
    vim.notify("[yapi-nvim] Buffer has no file name", vim.log.levels.ERROR)
    return
  end

  if not filepath:match("%.yapi$") and
     not filepath:match("%.yapi%.yml$") and
     not filepath:match("%.yapi%.yaml$")
  then
    vim.notify("[yapi-nvim] Not a yapi config file", vim.log.levels.WARN)
    return
  end

  if vim.bo.modified then
    vim.cmd("write")
  end

  local _, buf = open_result_window()
  vim.api.nvim_buf_set_option(buf, "modifiable", true)
  vim.api.nvim_buf_set_lines(buf, 0, -1, false, { "Running yapi..." })

  local cmd = { "yapi", "-c", filepath }

  vim.fn.jobstart(cmd, {
    stdout_buffered = true,
    stderr_buffered = true,

    on_stdout = function(_, data)
      if not data then return end
      vim.schedule(function()
        if vim.api.nvim_buf_is_valid(buf) then
          vim.api.nvim_buf_set_option(buf, "modifiable", true)
          vim.api.nvim_buf_set_lines(buf, 0, -1, false, data)
          vim.api.nvim_buf_set_option(buf, "modifiable", false)
        end
      end)
    end,

    on_stderr = function(_, data)
      if not data then return end
      vim.schedule(function()
        if vim.api.nvim_buf_is_valid(buf) then
          vim.api.nvim_buf_set_option(buf, "modifiable", true)
          local existing = vim.api.nvim_buf_get_lines(buf, 0, -1, false)
          vim.list_extend(existing, { "", "stderr:", unpack(data) })
          vim.api.nvim_buf_set_lines(buf, 0, -1, false, existing)
          vim.api.nvim_buf_set_option(buf, "modifiable", false)
        end
      end)
    end,

    on_exit = function(_, code)
      vim.schedule(function()
        if code ~= 0 then
          vim.notify("[yapi-nvim] yapi exited with code " .. code, vim.log.levels.ERROR)
        end
      end)
    end,
  })
end

function M.run()
  run_yapi_for_current()
end

function M.setup(opts)
  opts = opts or {}
  local hot_reload = opts.hot_reload ~= false

  vim.api.nvim_create_user_command("YapiRun", function()
    run_yapi_for_current()
  end, { desc = "Run yapi for current *.yapi file" })

  if hot_reload then
    local augroup = vim.api.nvim_create_augroup("YapiNvim", { clear = true })
    vim.api.nvim_create_autocmd("BufWritePost", {
      group = augroup,
      pattern = { "*.yapi", "*.yapi.yml", "*.yapi.yaml" },
      callback = function()
        run_yapi_for_current()
      end,
      desc = "Auto-run yapi on save",
    })
  end
end

return M

