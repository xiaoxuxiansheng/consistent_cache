package redis

const (
	// 通过 lua 脚本确保在 disable key 不存在时，才执行 key value 对写入
	LuaCheckEnableAndWriteCache = `
	local disable_key = KEYS[1];
	local disable_flag = redis.call("get",disable_key);
	if disable_flag then
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
