package consistent_cache

const (
	LuaCheckEnableAndWriteCache = `
	local enable_key = KEYS[1];
	local enable_flag = redis.call("get",enable_key);
	if enable_flag then
	    return 0;
	end
	local key = KEYS[2];
	local value = ARGV[1];
	redis.call("set",key,value);
	local cache_expire_seconds = tonumber(ARGV[2]);
	redis.call("expire",key,cache_expire_seconds);
	return 1;
`
)
