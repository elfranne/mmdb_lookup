package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/netip"
	"strconv"
	"strings"

	"github.com/elfranne/geoip2-golang/v2"
	"github.com/martinlindhe/base36"
)

func route(dbOpen *geoip2.Reader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := strings.TrimPrefix(r.URL.Path, "/")

		if ip, err := netip.ParseAddr(data); err != nil {
			mmdbLookup(dbOpen, &ip)
		}

		if value, err := strconv.ParseUint(data, 10, 64); err == nil {
			fmt.Fprintf("%q is a valid uint64 value: %d", data, value)
			return 
		}
		fmt.Fprintf(w, "%s\n", data)
		return 

	}
}

func mmdbLookup(dbOpen *geoip2.Reader, ip *netip.Addr) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		record, err := dbOpen.City(*ip)
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
		fmt.Fprintf(w, "%v\n", base36.Decode(record.Country.ISOCode))
	}
}

func decode(w http.ResponseWriter, r *http.Request) {
	decoded, err := strconv.ParseUint(strings.TrimPrefix(r.URL.Path, "/decode/"), 10, 64)
	if err != nil {
		http.Error(w, "06", http.StatusServiceUnavailable)
		fmt.Println(r.URL.Path + " failed : " + err.Error())
		return
	}
	fmt.Fprintf(w, "%v\n", base36.Encode(decoded))

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

	http.HandleFunc("/", route(dbOpen))
	fmt.Println("Listening on http://" + listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}
