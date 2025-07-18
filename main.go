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

		ipStr := strings.TrimPrefix(r.URL.Path, "/")
		ip, err := netip.ParseAddr(ipStr)
		if err != nil {
			http.Error(w, "00", http.StatusServiceUnavailable)
			fmt.Println(ip.String() + ": " + err.Error())
			return
		}

		record, err := dbOpen.City(ip)
		if err != nil {
			http.Error(w, "01", http.StatusServiceUnavailable)
			fmt.Println(ip.String() + ": " + err.Error())
			return
		}

		if !record.HasData() {
			http.Error(w, "03", http.StatusNotFound)
			fmt.Println(ip.String() + ": no data found")
			return
		}
		fmt.Fprintf(w, "%s\n", record.Country.ISOCode)
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
		log.Println(err.Error())
		return
	}
	defer dbOpen.Close()

	http.HandleFunc("/", mmdbLookup(dbOpen))
	fmt.Println("Listening on http://" + listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}
