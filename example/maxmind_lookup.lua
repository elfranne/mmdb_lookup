-- Source: https://github.com/elfranne/mmdb_lookup
-- Copyright (C) 2025 Elfranne
-- License: MIT

local function http_request(src)
  local s = core.tcp()
  s:connect("127.0.0.1:10000")
  s:send("GET /" .. src .. " HTTP/1.1\r\nHost: 127.0.0.1:10000\r\n\r\n")
  local msg = s:receive("*l")
  return msg

end

local function mmdb_lookup(txn)
  txn:set_var('txn.geoip_country', http_request(txn.f:src()))
end

core.register_action('mmdb_lookup', { 'tcp-req', 'http-req' }, mmdb_lookup, 0)
