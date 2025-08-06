# mmdb_lookup

## !!! WORK IN PROGRESS !!!

Originaly inspired by: <https://github.com/O-X-L/mmdb_lookup>

This is a minimalistic version with only country in order to do GEOIP in Haproxy.

The provided [lua example](./example/haproxy.conf) will query the webserver and store the value in the `txn.geoip_country` [variable](https://www.haproxy.com/documentation/haproxy-configuration-manual/latest/#2.8)

Hapoxy does not expose the sticky-tables in their stats pages or the built-in prometheus exporter, but you can use [haproxy_runtimeapi_exporter](https://github.com/elfranne/haproxy_runtimeapi_exporter) to collect metrics.

## Possible errors

err(1): failed to parse request, bad ip.(http 503)

err(2): failed to find ip in the provided mmdb. (http 404)

err(3): ip has no data in the provided mmdb. (http 404)

err(4): lua request failed. See in the [example](./example/haproxy.conf).

## Private and Local link

Private and local link will return `RFC1918`.

## IPv6

IPv6 should work but I have not done much testing.

## Open databases

You can use the Geo Open database ([more info](https://data.public.lu/en/datasets/geo-open-ip-address-geolocation-per-country-in-mmdb-format/))

## Example

```shell
wget https://cra.circl.lu/opendata/geo-open/mmdb-country/latest.mmdb -O ./mmdb-country.mmdb
go run main.go --mmdb ./mmdb-country.mmdb
curl "http://127.0.0.1:10000/192.168.0.1"
# RFC1918
curl "http://127.0.0.1:10000/1.1.1.1"
# US
curl "http://127.0.0.1:10000/2606:4700:4700::1111"
# US
```
