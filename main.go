package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	geoip2 "github.com/oschwald/geoip2-golang"
	"github.com/tidwall/gjson"
)

type _IPLocation struct {
	lat string
	lon string
}

type weatherInfo struct {
	City        string
	Country     template.HTML
	Date        string
	Temperature string
	Weather     string
	Code        string
	Longitude   string
	Latitude    string
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

func getUserIp(req *http.Request) string {
	Ipstring := req.RemoteAddr
	index := strings.Index(Ipstring, ":")
	if index > 0 {
		Ipstring = Ipstring[:index]
	}

	fmt.Println(Ipstring)

	if proxies := req.Header.Get("x-forwarded-for"); proxies != "" {
		ips := strings.Split(proxies, ", ")
		return ips[0]
	}

	return Ipstring
}

func getIPLocation(req *http.Request) _IPLocation {
	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// If you are using strings that may be invalid, check that ip is not nil
	record, err := db.City(net.ParseIP("80.104.158.156" /* getUserIp(req) */))
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
	// Template Creation
	w.Header().Add("Content Type", "text/html")

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

	dateStr := date.String()
	codeStr := code.String()
	tmplFiles := getTemplateFiles(dateStr, codeStr)
	tmpl := template.Must(template.ParseFiles(tmplFiles...))
	// templates := template.Must(template.ParseGlob("templates/*.html"))

	context := weatherInfo{city.String(),
		template.HTML("<i>" + country.String() + "</i>"),
		dateStr,
		temperature.String(),
		weather.String(),
		codeStr,
		lon, lat}

	/*
		http://stackoverflow.com/questions/24755509/golang-templates-using-conditions-inside-templates
		http://stackoverflow.com/questions/24755509/golang-templates-using-conditions-inside-templates
		http://stackoverflow.com/questions/24755509/golang-templates-using-conditions-inside-templates
	*/
	if err := tmpl.ExecuteTemplate(w, "main", context); err != nil {
		return err
	}

	return nil
}

func getHourFromDateQuery(date string) int {
	index := strings.Index(date, ":")

	/* TODO: test with fmt.Println */

	if index > 0 {
		hour, err := strconv.Atoi(date[index-2 : index])
		// meridiem := date[index+4 : index+6]
		if date[index+4:index+6] == "PM" {
			hour += 12
		}

		if err != nil {
			panic(err)
		}

		return hour
	}

	return -1
}

func getTemplateFiles(dateStr string, codeStr string) []string {
	hour := getHourFromDateQuery(dateStr)

	tmplFiles := []string{
		"templates/index.html",
		"templates/content.html"}

	if hour > 21 || hour < 6 {
		tmplFiles = append(tmplFiles, "templates/night_theme.html")
	} else {
		tmplFiles = append(tmplFiles, "templates/day_theme.html")
	}

	// appending a weather svg icon depending on the resulting code from the query
	codeInt, _ := strconv.Atoi(codeStr)
	if codeInt < 25 {
		tmplFiles = append(tmplFiles, "icons/storm.svg")
	} else if codeInt < 30 {
		tmplFiles = append(tmplFiles, "icons/cloudy.svg")
	} else if codeInt <= 34 {
		tmplFiles = append(tmplFiles, "icons/sunny.svg")
	} else if codeInt == 35 {
		tmplFiles = append(tmplFiles, "icons/storm.svg")
	} else if codeInt == 36 {
		tmplFiles = append(tmplFiles, "icons/sunny.svg")
	} else {
		tmplFiles = append(tmplFiles, "icons/storm.svg")
	}

	return tmplFiles
}
