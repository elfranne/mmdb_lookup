package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/netip"
	"strings"

	"github.com/oschwald/geoip2-golang/v2"
)

func mmdbLookup(w http.ResponseWriter, r *http.Request) {
	ipStr := strings.TrimPrefix(r.URL.Path, "/")

	dbOpen, err := geoip2.Open(flag.Lookup("mmdb").Value.(flag.Getter).Get().(string))
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer dbOpen.Close()

	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	record, err := dbOpen.City(ip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	if !record.HasData() {
		http.Error(w, "No data found for this IP", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "%s\n", record.Country.ISOCode)
}

func main() {
	var listen string
	var mmdb string

	flag.StringVar(&listen, "l", "127.0.0.1:10000", "Address and port to listen on (default: 127.0.0.1:10000)")
	flag.StringVar(&mmdb, "mmdb", "./mmdb-country.mmdb", "Path to the mmdb (default: \"./mmdb-country.mmdb\")")
	flag.Parse()

	http.HandleFunc("/", mmdbLookup)
	fmt.Println("Listening on http://" + listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}
