package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"

	"github.com/oschwald/maxminddb-golang"
)

var DB_CITY = "/var/lib/GeoIP/GeoLite2-City.mmdb"
var DB_ASN = "/var/lib/GeoIP/GeoLite2-ASN.mmdb"
var RETURN_PLAIN = false

var DB_TYPE = DB_TYPE_MAXMIND

const DB_TYPE_MAXMIND uint8 = 2

// MaxMind schema: https://github.com/maxmind/MaxMind-DB/tree/main/source-data
var MAXMIND_COUNTRY struct {
	Country struct {
		Code         string `maxminddb:"iso_code"`
		Id           uint   `maxminddb:"geoname_id"`
		EuopeanUnion bool   `maxminddb:"is_in_european_union"`
	} `maxminddb:"country"`
	RegisteredCountry struct {
		Code         string `maxminddb:"iso_code"`
		Id           uint   `maxminddb:"geoname_id"`
		EuopeanUnion bool   `maxminddb:"is_in_european_union"`
	} `maxminddb:"registered_country"`
	Continent struct {
		Code  string            `maxminddb:"code"`
		Id    uint              `maxminddb:"geoname_id"`
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"continent"`
}

var MAXMIND_ASN struct {
	ASN  string `maxminddb:"autonomous_system_number"`
	Name string `maxminddb:"autonomous_system_organization"`
}

var MAXMIND_CITY struct {
	City struct {
		Code string `maxminddb:"iso_code"`
		Id   uint   `maxminddb:"geoname_id"`
	} `maxminddb:"country"`
	Country struct {
		Code         string `maxminddb:"iso_code"`
		Id           uint   `maxminddb:"geoname_id"`
		EuopeanUnion bool   `maxminddb:"is_in_european_union"`
	} `maxminddb:"country"`
	RegisteredCountry struct {
		Code         string `maxminddb:"iso_code"`
		Id           uint   `maxminddb:"geoname_id"`
		EuopeanUnion bool   `maxminddb:"is_in_european_union"`
	} `maxminddb:"registered_country"`
	Continent struct {
		Code  string            `maxminddb:"code"`
		Id    uint              `maxminddb:"geoname_id"`
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"continent"`
	Location struct {
		AccuracyRadius uint    `maxminddb:"accuracy_radius"`
		Latitude       float32 `maxminddb:"latitude"`
		Longitude      float32 `maxminddb:"longitude"`
		Timezone       string  `maxminddb:"time_zone"`
	} `maxminddb:"location"`
	Postal struct {
		Code string `maxminddb:"code"`
	} `maxminddb:"postal"`
	Traits struct {
		IsAnycast        bool `maxminddb:"is_anycast"`
		IsAnonymousProxy bool `maxminddb:"is_anonymous_proxy"`
	} `maxminddb:"traits"`
}

var MAXMIND_PRIVACY struct { // also called 'anonymous'
	Any          bool `maxminddb:"is_anonymous"`
	Vpn          bool `maxminddb:"is_anonymous_vpn"`
	Tor          bool `maxminddb:"is_tor_exit_node"`
	Hosting      bool `maxminddb:"is_hosting_provider"`
	PublicProxy  bool `maxminddb:"is_public_proxy"`
	PrivateProxy bool `maxminddb:"is_residential_proxy"`
}

var FUNC_MAPPING = map[uint8]interface{}{
	DB_TYPE_MAXMIND: map[string]interface{}{
		"city": MaxMindCity,
		"asn":  MaxMindAsn,
	},
}

var FUNC = FUNC_MAPPING[DB_TYPE].(map[string]interface{})

func GetMapValue(dataStructure interface{}, name string) interface{} {
	return reflect.Indirect(
		reflect.ValueOf(&dataStructure),
	).Elem().Interface().(map[string]interface{})[name]
}

func LogError(prefix string, err interface{}) {
	log.Fatalf("%v, Error: %v", prefix, err)
}

func errorResponse(w http.ResponseWriter, m string) {
	w.WriteHeader(http.StatusBadRequest)
	_, err := io.WriteString(w, fmt.Sprintf("%v\n", m))
	if err != nil {
		log.Fatal(err)
	}
}

func returnResult(w http.ResponseWriter, data interface{}, logPrefix string) {
	if RETURN_PLAIN {
		w.Header().Set("Content-Type", "text/plain")
		_, err := io.WriteString(w, fmt.Sprintf("%+v\n", data))
		if err != nil {
			LogError(logPrefix, err)
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			LogError(logPrefix, err)
			errorResponse(w, "Failed to JSON-encode data")
		}
	}
}

func getClientIP(r *http.Request) (string, error) {
	fwdIPs := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
	if len(fwdIPs) > 0 {
		netIP := net.ParseIP(fwdIPs[len(fwdIPs)-1])
		if netIP != nil {
			return netIP.String(), nil
		}
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		netIP := net.ParseIP(realIP)
		if netIP != nil {
			return netIP.String(), nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	netIP := net.ParseIP(ip)
	if netIP != nil {
		ip := netIP.String()
		if ip == "::1" {
			return "127.0.0.1", nil
		}
		return ip, nil
	}

	return "", errors.New("IP not found")
}

func MaxMindCity(ip net.IP) (interface{}, error) {
	return lookupBase(ip, MAXMIND_CITY, DB_CITY)
}

func MaxMindAsn(ip net.IP) (interface{}, error) {
	return lookupBase(ip, MAXMIND_ASN, DB_ASN)
}

func lookupBase(ip net.IP, dataStructure interface{}, dbFile string) (interface{}, error) {
	db, err := maxminddb.Open(dbFile)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.Lookup(ip, &dataStructure)
	if err != nil {
		return nil, err
	}
	return dataStructure, nil
}

func geoIpLookup(w http.ResponseWriter, r *http.Request) {
	ipStr := r.URL.Query().Get("ip")
	lookupStr := r.URL.Query().Get("lookup")
	filterStr := r.URL.Query().Get("filter")
	logPrefix := fmt.Sprintf("IP: '%v', Lookup: '%v', Filter: '%v'", ipStr, lookupStr, filterStr)

	if ipStr == "" {
		clientIpStr, err := getClientIP(r)
		if err == nil {
			ipStr = clientIpStr
		}
	}

	if lookupStr == "" || ipStr == "" {
		errorResponse(w, "Either 'lookup' or 'ip' were not provided")
		return
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		errorResponse(w, "Invalid IP provided")
		return
	}

	data, err := FUNC[lookupStr].(func(net.IP) (interface{}, error))(ip)
	if data == nil {
		errorResponse(w, "Invalid LOOKUP provided")
		return
	}
	if err != nil {
		LogError(logPrefix, err)
		errorResponse(w, "Failed to lookup data")
		return
	}

	if filterStr != "" {
		filteredData := data
		for _, subFilterStr := range strings.Split(filterStr, ".") {
			defer func() {
				if err := recover(); err != nil {
					LogError(logPrefix, err)
					errorResponse(w, "Invalid FILTER provided")
				}
			}()
			filteredData = GetMapValue(filteredData, subFilterStr)
			if filteredData == nil {
				errorResponse(w, "Invalid FILTER provided")
				return
			}
		}
		returnResult(w, filteredData, logPrefix)
		return
	}

	returnResult(w, data, logPrefix)
}

func httpServer(listenAddr string, listenPort uint) {
	http.HandleFunc("/", geoIpLookup)
	var listenStr = fmt.Sprintf("%v:%v", listenAddr, listenPort)
	fmt.Println("Listening on http://" + listenStr)
	log.Fatal(http.ListenAndServe(listenStr, nil))
}

func main() {
	var listenAddr string
	var listenPort uint

	flag.StringVar(&listenAddr, "l", "127.0.0.1", "Address to listen on")
	flag.UintVar(&listenPort, "p", 10000, "Port to listen on")

	flag.StringVar(&DB_CITY, "city", DB_CITY, "Path to the city-database (optional)")
	flag.StringVar(&DB_ASN, "asn", DB_ASN, "Path to the asn-database (optional)")
	flag.BoolVar(&RETURN_PLAIN, "plain", RETURN_PLAIN, "If the result should be returned in plain text format")
	flag.Parse()

	httpServer(listenAddr, listenPort)
}
