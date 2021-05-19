-- Clear database and generate a set of keys
-- The number of keys in the set is specified using the first argument passed
-- to the script.
redis.call("FLUSHDB")
for i=1,KEYS[1] do
    local n = string.format("%.16d", i)
    redis.call("SET", n, n)
end

return "Ok!"
