# maxmind-lookup

Originaly inspired by <https://github.com/O-X-L/geoip-lookup-service>

This is a minimalistic version to just return country in order to do GEOIP blocking in Haproxy.

Using test database:

```shell
wget https://raw.githubusercontent.com/maxmind/MaxMind-DB/refs/heads/main/test-data/GeoIP2-Country-Test.mmdb -o ./test/GeoIP2-Country-Test.mmdb
go run main.go --mmdb ./test/GeoIP2-Country-Test.mmdb
curl "http://127.0.0.1:10000/81.2.69.192"
```


