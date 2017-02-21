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
	"github.com/tidwall/gjson"
)

func main() {
	serveFile := http.StripPrefix("/res/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/", indexHandler)
	http.Handle("/res/", serveFile)
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

	/*city := record.City.Names["en"]
	subdiv := record.Subdivisions[0].Names["en"]
	country := record.Country.Names["en"]*/
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

	yahoores, err := http.Get(Url.String())
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(yahoores.Body)
	/*city = queryResult.results.channel.location.city;
	  country = queryResult.results.channel.location.country;
	  date = queryResult.results.channel.item.condition.date;
	  temperature = queryResult.results.channel.item.condition.temp;
	  weather = queryResult.results.channel.item.condition.text;
	  lat = queryResult.results.channel.item.lat;
	  lon = queryResult.results.channel.item.long;
	  var condition_code = queryResult.results.channel.item.condition.code;*/
	result := gjson.GetBytes(body, "query.results.channel.location.city")
	fmt.Println(result)
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content Type", "text/html")

	tmpl, err := template.ParseFiles("index.html")

	if err != nil {
		fmt.Println(err)
		return
	}

	getWeatherInfo(w, req)

	if err := tmpl.ExecuteTemplate(w, "index.html", nil); err != nil {
		fmt.Println(err)
		return
	}
}
