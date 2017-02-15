package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/oschwald/geoip2-golang"
)

func main() {
	/*
	   select * from weather.forecast where woeid in (SELECT woeid FROM geo.places WHERE text="({lat},{lon})")

	       'http://query.yahooapis.com/v1/public/yql?q='
	                    + encodedQuery + '&format=json

	   http://query.yahooapis.com/v1/public/yql?q=select%20*%20from%20weather.forecast%20where%20woeid%20in%20(SELECT%20woeid%20FROM%20geo.places%20WHERE%20text%3D%22(45%2C10)%22)&format=json

	   e fai quello che vuoi col risultato in json
	*/

	serveFile := http.StripPrefix("/res/", http.FileServer(http.Dir("./")))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/getweatherinfo", getWeatherInfo)
	http.Handle("/res", serveFile)
	http.ListenAndServe(":8080", nil)
}

func getWeatherInfo(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content Type", "application/json")

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

	yahooquery := "select * from weather.forecast where woeid in (SELECT woeid FROM geo.places WHERE text=\"(" +
		lat + "," + lon + ")\")"

	Url, err := url.Parse("http://query.yahooapis.com/")
	if err != nil {
		fmt.Printf("url parsing failed")
	}

	parameters := url.Values{}
	parameters.Add("q", yahooquery)
	parameters.Add("format", "json")

	Url.Path += "/v1/public/yql"
	Url.RawQuery = parameters.Encode()
	//fmt.Printf("Encoded URL is %q\n", Url.String())

	yahoores, err := http.Get(Url.String())

	fmt.Fprint(w, "hello "+city+" "+subdiv+" "+country+" "+lat+" "+lon+"  ")
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(yahoores.Body)
	fmt.Fprint(w, string(body))
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content Type", "text/html")

	tmpl, err := template.ParseFiles("index.html")

	if err != nil {
		fmt.Println(err)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "index.html", nil); err != nil {
		fmt.Println(err)
		return
	}
}
