# mmdb_lookup

## WORK IN PROGRESS

Originaly inspired by: <https://github.com/O-X-L/mmdb_lookup>

This is a minimalistic version with only country in order to do GEOIP blocking in Haproxy.

You can use the Geo Open database ([more info](https://data.public.lu/en/datasets/geo-open-ip-address-geolocation-per-country-in-mmdb-format/)):

```shell
wget https://cra.circl.lu/opendata/geo-open/mmdb-country/latest.mmdb -O ./mmdb-country.mmdb
go run main.go --mmdb ./mmdb-country.mmdb
curl "http://127.0.0.1:10000/192.168.0.1"
```
