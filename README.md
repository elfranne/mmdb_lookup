# maxmind-lookup

Originaly inspired by https://github.com/O-X-L/geoip-lookup-service

This is a minimalistic version to just return country in order to do GEOIP blocking in Haproxy.

Using test database:

```shell
wget https://raw.githubusercontent.com/maxmind/MaxMind-DB/refs/heads/main/test-data/GeoIP2-Country-Test.mmdb -o ./test/GeoIP2-Country-Test.mmdb
go run main.go --mmdb ./test/GeoIP2-Country-Test.mmdb
curl "http://127.0.0.1:10000/81.2.69.192"
```

Lua sript to load in Haproxy:

```lua
-- Source: https://github.com/elfranne/maxmind-lookup
-- Copyright (C) 2025 Elfranne
-- License: MIT

-- NOTE: the ltrim parameter can be used to remove a prefix - like: 'AS1337' => '1337'

local function http_request(lookup, filter, src, ltrim)
    local s = core.tcp()

    local addr = '127.0.0.1'
    local port = 10000

    local hdrs = {
        [1] = string.format('host: %s:%s', addr, port),
        [2] = 'accept: */*',
        [3] = 'connection: close'
    }

    local req = {
        [1] = 'GET /?lookup=' .. lookup .. '&ip=' .. src .. '&filter=' .. filter .. ' HTTP/1.1',
        [2] = table.concat(hdrs, '\r\n'),
        [3] = '\r\n'
    }

    req = table.concat(req, '\r\n')

    s:connect(addr, port)
    s:send(req)
    while true do
        local line = s:receive('*l')
        if not line then break end
        if line == '' then break end
    end
    local res_body = s:receive('*a')
    if res_body == nil then
        return '00'
    end
    return string.sub(res_body, 1 + ltrim, -2)
end

-- examples for MaxMind:

local function lookup_geoip_country(txn)
    country_code = http_request('country', 'country.iso_code', txn.f:src(), 0)
    txn:set_var('txn.geoip_country', country_code)
end

local function lookup_geoip_asn(txn)
    asn = http_request('asn', 'autonomous_system_number', txn.f:src(), 0)
    txn:set_var('txn.geoip_asn', asn)
end

local function lookup_geoip_asname(txn)
    asname = http_request('asn', 'autonomous_system_organization', txn.f:src(), 0)
    txn:set_var('txn.geoip_asname', asname)
end

-- examples for IPInfo:

local function lookup_geoip_country(txn)
    country_code = http_request('country', 'country', txn.f:src(), 0)
    txn:set_var('txn.geoip_country', country_code)
end

local function lookup_geoip_asn(txn)
    asn = http_request('asn', 'asn', txn.f:src(), 2)
    txn:set_var('txn.geoip_asn', asn)
end

local function lookup_geoip_asname(txn)
    asname = http_request('asn', 'name', txn.f:src(), 0)
    txn:set_var('txn.geoip_asname', asname)
end

-- examples end

core.register_action('lookup_geoip_country', {'tcp-req', 'http-req'}, lookup_geoip_country, 0)
core.register_action('lookup_geoip_asn', {'tcp-req', 'http-req'}, lookup_geoip_asn, 0)
core.register_action('lookup_geoip_asname', {'tcp-req', 'http-req'}, lookup_geoip_asname, 0)
```

Haproxy config:

```conf
WIP
```
