-- Global Neovim client for the justify CLI.
-- Save as: ~/.config/nvim/lua/justify_client.lua  (this file is embedded by justify and can be written via `justify editor --editor neovim-global`)

local M = {}

M.opts = {
	justify_cmd = "auto", -- "auto" (prefer ./justify) or an absolute/path or "justify" on PATH
	root_patterns = { ".git", ".justify/targets.json", "Cargo.toml", "go.mod", "package.json", "CMakeLists.txt" },
	default_node_port = 9229,
}

local function is_windows()
	return vim.loop.os_uname().version:match("Windows")
end

local function project_root()
	local cwd = vim.fn.getcwd()
	for _, marker in ipairs(M.opts.root_patterns) do
		local found = vim.fs.find(marker, { upward = true, path = cwd })[1]
		if found then return vim.fs.dirname(found) end
	end
	return cwd
end

local function justify_cmd_for(root)
	if M.opts.justify_cmd == "auto" then
		local cand = root .. "/justify" .. (is_windows() and ".exe" or "")
		if vim.fn.filereadable(cand) == 1 or vim.fn.executable(cand) == 1 then
			return cand
		end
		return "justify"
	end
	return M.opts.justify_cmd
end

local function system_ok(cmd, cwd)
	local out = vim.fn.systemlist(cmd, cwd)
	return vim.v.shell_error == 0, out
end

local function json_decode(s)
	local ok, res = pcall(vim.json.decode, s)
	if ok then return res end
	return nil
end

local function read_targets(root)
	local bin = justify_cmd_for(root)
	local ok, out = system_ok(bin .. " targets --json", root)
	if not ok then return {} end
	local data = json_decode(table.concat(out, "\n"))
	if type(data) ~= "table" then return {} end
	for _, t in ipairs(data) do
		t.args    = t.args or {}
		t.env     = t.env or {}
		t.cwd     = t.cwd or ""
		t.request = t.request or ""
	end
	return data
end

local function program_for_last(root)
	local bin = justify_cmd_for(root)
	local ok, out = system_ok(bin .. " program", root)
	if not ok or #out == 0 then return nil end
	return out[1]
end

local function set_last(root, name)
	local bin = justify_cmd_for(root)
	system_ok(bin .. " target set " .. name, root)
end

local function ensure_adapters()
	local dap = require("dap")
	if not dap.adapters.codelldb then
		dap.adapters.codelldb = {
			type = "server",
			port = "${port}",
			executable = { command = vim.fn.stdpath("data") .. "/mason/bin/codelldb" },
		}
	end
end

local function effective_cwd(tgt, root)
	if tgt.cwd and tgt.cwd ~= "" then
		if vim.fn.isabsolute(tgt.cwd) == 1 then return tgt.cwd end
		return vim.fn.fnamemodify(root .. "/" .. tgt.cwd, ":p:h")
	end
	return root
end

local function env_table(env)
	local out = {}
	if type(env) == "table" then
		for k, v in pairs(env) do out[k] = v end
	end
	return out
end

local function run_target(tgt, root)
	ensure_adapters()
	local dap  = require("dap")
	local kind = (tgt.kind or ""):lower()
	local req  = (tgt.request or ""):lower()
	local prog = tgt.program
	local cwd  = effective_cwd(tgt, root)
	local args = tgt.args or {}
	local env  = env_table(tgt.env)

	if kind == "node" then
		if prog == "attach:node" or req == "attach" then
			dap.run({
				type = "pwa-node",
				request = "attach",
				name = "Node Attach - " .. tgt.name,
				cwd = cwd,
				port = tgt.port or M.opts.default_node_port,
			})
		else
			dap.run({
				type = "pwa-node",
				request = "launch",
				name = "Node Launch - " .. tgt.name,
				cwd = cwd,
				program = prog,
				args = args,
				env = env,
				skipFiles = { "<node_internals>/**" },
			})
		end
		return
	end

	if kind == "go" then
		dap.run({
			type = "go",
			request = "launch",
			name = "Go Launch - " .. tgt.name,
			program = "${workspaceFolder}",
			cwd = cwd,
			args = args,
			env = env,
		})
		return
	end

	-- rust/cpp/native
	dap.run({
		type = "codelldb",
		request = "launch",
		name = "Native Launch - " .. tgt.name,
		program = prog,
		cwd = cwd,
		args = args,
		env = env,
		stopOnEntry = false,
	})
end

function M.pick_and_debug()
	local root = project_root()
	local targets = read_targets(root)
	if #targets == 0 then
		vim.notify("[justify] No targets. Run `justify targets --init` and edit .justify/targets.json",
			vim.log.levels.ERROR)
		return
	end
	local items = {}
	for _, t in ipairs(targets) do
		local meta = (t.request and t.request ~= "" and (" (" .. t.request .. ")") or "")
		table.insert(items, string.format("%s  [%s]%s  %s", t.name, t.kind, meta, t.program))
	end
	vim.ui.select(items, { prompt = "Select target" }, function(_, idx)
		if not idx then return end
		local t = targets[idx]
		set_last(root, t.name)
		run_target(t, root)
	end)
end

function M.debug_last()
	local root = project_root()
	local prog = program_for_last(root)
	if not prog then
		vim.notify("[justify] No last-used target. Use :JustPickTarget", vim.log.levels.WARN)
		return
	end
	local targets = read_targets(root)
	local tgt = nil
	for _, t in ipairs(targets) do
		if t.name == prog or t.program == prog then
			tgt = t; break
		end
	end
	if not tgt then
		tgt = { name = "last-used", kind = "cpp", program = prog, cwd = "" }
	end
	run_target(tgt, root)
end

function M.diag()
	local root = project_root()
	local ts = read_targets(root)
	local prog = program_for_last(root)
	local lines = {
		"root: " .. root,
		"targets: " .. tostring(#ts),
	}
	for _, t in ipairs(ts) do
		table.insert(lines,
			string.format("  - %s [%s]%s  %s  cwd=%s  port=%s", t.name, t.kind or "?",
				t.request and ("/" .. t.request) or "", t.program or "?", t.cwd or ".", t.port or ""))
	end
	table.insert(lines, "last-used program: " .. (prog or "<none>"))
	vim.notify(table.concat(lines, "\n"), vim.log.levels.INFO, { title = "justify diag" })
end

function M.setup_commands()
	vim.api.nvim_create_user_command("JustPickTarget", M.pick_and_debug, {})
	vim.api.nvim_create_user_command("JustDebug", M.debug_last, {})
	vim.api.nvim_create_user_command("JustifyDiag", M.diag, {})
end

return M
