package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/netip"
	"strings"

	"github.com/elfranne/geoip2-golang/v2"
)

func mmdbLookup(dbOpen *geoip2.Reader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := strings.TrimPrefix(r.URL.Path, "/")
		ip, err := netip.ParseAddr(data)
		if err != nil {
			log.Printf("err(1): failed to parse: %s\n", data)
			http.Error(w, "1\n", http.StatusServiceUnavailable)
			return
		}

		record, err := dbOpen.City(ip)
		if err != nil {
			log.Printf("err(2) %s: %s\n", ip.String(), err.Error())
			http.Error(w, "2\n", http.StatusNotFound)
			return
		}

		if ip.IsPrivate() || ip.IsLoopback() {
			fmt.Fprintf(w, "%v\n", "RFC1918")
			return
		}

		if !record.HasData() {
			log.Printf("err(3) No data found for IP %s\n", ip.String())
			http.Error(w, "3\n", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "%v\n", record.Country.ISOCode)
	}
}

func main() {
	var listen string
	var mmdb string

	flag.StringVar(&listen, "l", "127.0.0.1:10000", "Address and port to listen on (default: 127.0.0.1:10000)")
	flag.StringVar(&mmdb, "mmdb", "./mmdb-country.mmdb", "Path to the mmdb (default: \"./mmdb-country.mmdb\")")
	flag.Parse()

	dbOpen, err := geoip2.Open(mmdb)
	if err != nil {
		log.Fatalf("Error opening MMDB: %s\n", err.Error())
		return
	}
	defer dbOpen.Close()

	http.HandleFunc("/", mmdbLookup(dbOpen))
	log.Printf("Listening on http://%s\n", listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}
