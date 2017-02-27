package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"

	geoip2 "github.com/oschwald/geoip2-golang"
	"github.com/tidwall/gjson"
)

type _IPLocation struct {
	lat string
	lon string
}

type WeatherInfo struct {
	City string
}

func main() {
	/*
			example query
		   select * from weather.forecast where woeid in (SELECT woeid FROM geo.places WHERE text="({lat},{lon})")

		       'http://query.yahooapis.com/v1/public/yql?q='
		                    + encodedQuery + '&format=json

		   http://query.yahooapis.com/v1/public/yql?q=select%20*%20from%20weather.forecast%20where%20woeid%20in%20(SELECT%20woeid%20FROM%20geo.places%20WHERE%20text%3D%22(45%2C10)%22)&format=json
	*/
	serveFile := http.StripPrefix("/res/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/", getWeatherInfo)
	http.Handle("/res/", serveFile)
	http.ListenAndServe(":8080", nil)
}

func getWeatherInfo(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content Type", "application/json")

	location := getIPLocation(req)
	response, err := queryYahooWeatherAPI(location.lat, location.lon)
	if err != nil {
		panic(err)
	}

	if err = constructResponse(w, response, location); err != nil {
		panic(err)
	}
}

func getIPLocation(req *http.Request) _IPLocation {
	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// If you are using strings that may be invalid, check that ip is not nil
	record, err := db.City(net.ParseIP("80.104.158.156"))
	if err != nil {
		log.Fatal(err)
	}

	/*
		we could also gather city location directly from GeoLite's database

		city := record.City.Names["en"]
		subdiv := record.Subdivisions[0].Names["en"]
		country := record.Country.Names["en"]
	*/
	lat := strconv.FormatFloat(record.Location.Latitude, 'f', -1, 64)
	lon := strconv.FormatFloat(record.Location.Longitude, 'f', -1, 64)

	return _IPLocation{lat, lon}
}

func queryYahooWeatherAPI(lat string, lon string) (io.Reader, error) {
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

	return yahoores.Body, err
}

func constructResponse(w http.ResponseWriter, response io.Reader, iploc _IPLocation) error {
	body, err := ioutil.ReadAll(response)
	if err != nil {
		return err
	}

	city := gjson.GetBytes(body, "query.results.channel.location.city")
	country := gjson.GetBytes(body, "query.results.channel.location.country")
	date := gjson.GetBytes(body, "query.results.channel.item.condition.date")
	temperature := gjson.GetBytes(body, "query.results.channel.item.condition.temp")
	weather := gjson.GetBytes(body, "query.results.channel.item.condition.text")
	code := gjson.GetBytes(body, "query.results.channel.item.condition.code")
	lon := iploc.lon
	lat := iploc.lat

	var dat interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		return err
	}

	fmt.Println(city, country, date, temperature, code, weather, lon, lat)

	// Template Creation
	w.Header().Add("Content Type", "text/html")

	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		return err
	}

	context := WeatherInfo{City: city.String()}
	if err := tmpl.ExecuteTemplate(w, "index.html", context); err != nil {
		return err
	}

	return nil
}
