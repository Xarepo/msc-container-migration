-- Clear database and generate a set of keys
-- The number of keys in the set is specified using the first argument passed
-- to the script.
redis.call("FLUSHDB")
for i=1,KEYS[1] do
    redis.call("SET", i, i)
end

return "Ok!"
