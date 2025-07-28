  -- This lua script is adapted from:
  -- https://github.com/haproxytechblog/haproxy-lua-samples/blob/master/action/ipchecker/ipchecker.lua
  -- The contents of this file are Copyright (c) 2019. HAProxy Technologies. All Rights Reserved.
  -- Original license is GPLv2
  local function mmdb_lookup(txn)
      local addr = '127.0.0.1'
      local port = 10000
      local hdrs = {
          [1] = string.format('host: %s:%s', addr, port),
          [2] = 'accept: */*',
          [3] = 'connection: close'
      }
      local req = {
          [1] = string.format('GET /%s HTTP/1.1', tostring(txn.f:src())),
          [2] = table.concat(hdrs, '\r\n'),
          [3] = '\r\n'
      }
      req = table.concat(req,  '\r\n')
      local socket = core.tcp()
      socket:settimeout(5)
      if socket:connect(addr, port) then
          if socket:send(req) then
              -- Skip response headers
              while true do
                  local line, _ = socket:receive('*l')
                  if not line then break end
                  if line == '' then break end
              end
              local content = socket:receive('*a')
              if content then
                  txn:set_var('txn.geoip_country', content)
                  return
              end
          else
              core.Alert('Could not connect to mmdb_lookup (send)')
          end
          socket:close()
      else
          core.Alert('Could not connect to mmdb_lookup (connect)')
      end
      txn:set_var('txn.geoip_country', '4')
  end

  core.register_action('mmdb_lookup', {'http-req'}, mmdb_lookup, 0)