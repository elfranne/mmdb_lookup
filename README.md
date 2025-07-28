# mmdb_lookup

## !!! WORK IN PROGRESS !!!

Originaly inspired by: <https://github.com/O-X-L/mmdb_lookup>

This is a minimalistic version with only country in order to do GEOIP in Haproxy.
With Haproxy you can use sticky-tables to cache requests but it can only store integers (32bit positive) as General Purpose Tag (GPT).
`mmdb_lookup` will return countries (as [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)) in [Base36](https://en.wikipedia.org/wiki/Base36) encoding to be able to store the data in GPTs.

The provided [lua example](./example/haproxy.conf) will query the webserver and store the value in the `txn.geoip_country` [variable](https://www.haproxy.com/documentation/haproxy-configuration-manual/latest/#2.8)

You can also query a Base36 integer to get the human readable (see example below).

Hapoxy does not expose the sticky-tables in their stats pages or the built-in prometheus exporter, but you can use [haproxy_runtimeapi_exporter](https://github.com/elfranne/haproxy_runtimeapi_exporter) to collect metrics.

## Possible errors

err(1): failed to parse request, bad ip or not Base36.(http 503)

err(2): failed to find ip in the provided mmdb. (http 404)

err(3): ip has no data in the provided mmdb. (http 404)

err(4): lua request failed. See in the [example](./example/haproxy.conf).

## Private and Local link

Private and local link will return `1918` encoded in Base36 (`58364`), `RFC1918` was more than 32bits in Base36.

## IPv6

IPv6 should work but I have not done much testing.

## TODO
Ensure the output to be int32.

## Open databases

You can use the Geo Open database ([more info](https://data.public.lu/en/datasets/geo-open-ip-address-geolocation-per-country-in-mmdb-format/))

## Example

```shell
wget https://cra.circl.lu/opendata/geo-open/mmdb-country/latest.mmdb -O ./mmdb-country.mmdb
go run main.go --mmdb ./mmdb-country.mmdb
curl "http://127.0.0.1:10000/192.168.0.1"
# 58364
curl "http://127.0.0.1:10000/58364"
# 1918
curl "http://127.0.0.1:10000/1.1.1.1"
# 1108
curl "http://127.0.0.1:10000/2606:4700:4700::1111"
# 1108
curl "http://127.0.0.1:10000/1108"
# US
```
