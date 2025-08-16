-- [TODO] Implement error handling for file loading
function loadConfig(file)
	local config = {}

	-- [NOTE] Assuming file format is key=value
	for line in io.lines(file) do
		local key, value = line:match("^(.-)=(.-)$")
		config[key] = value
	end

	return config
end

-- [FIXME] Crashes if file doesn't exist
config = loadConfig("settings.ini")

-- [HACK] Global access for lazy reasons
GLOBAL_CONFIG = config

-- [OPTIMIZE] Cache config values for faster lookup

-- [BUG] Doesn't handle duplicate keys properly

-- [DEPRECATED] Use newConfigLoader() instead

