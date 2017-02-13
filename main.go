package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/oschwald/geoip2-golang"
)

func main() {
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	// w.Header().Add("Content Type", "text/html")

	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// If you are using strings that may be invalid, check that ip is not nil
	ip := net.ParseIP("80.104.158.156")
	record, err := db.City(ip)
	if err != nil {
		log.Fatal(err)
	}

	city := record.City.Names["en"]
	subdiv := record.Subdivisions[0].Names["en"]
	country := record.Country.Names["en"]
	lat := strconv.FormatFloat(record.Location.Latitude, 'f', -1, 64)
	lon := strconv.FormatFloat(record.Location.Longitude, 'f', -1, 64)

	fmt.Fprint(w, "hello "+city+" "+subdiv+" "+country+" "+lat+" "+lon)
}
