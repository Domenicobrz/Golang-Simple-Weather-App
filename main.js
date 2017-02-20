window.addEventListener("load", init);

var queryResult;
var date, temperature, weather, lat, lon, city, country;

function init() {
    "use strict";
    xhr = new XMLHttpRequest();
    xhr.addEventListener("load", reqHandler);
    xhr.open("GET", "getweatherinfo");
    xhr.send();
}

function reqHandler(evt) {
    "use strict";
    console.log(JSON.parse(evt.target.responseText));
    queryResult = evt.target.responseText;


    city = queryResult.results.channel.location.city;
    country = queryResult.results.channel.location.country;
    date = queryResult.results.channel.item.condition.date;
    temperature = queryResult.results.channel.item.condition.temp;
    weather = queryResult.results.channel.item.condition.text;
    lat = queryResult.results.channel.item.lat;
    lon = queryResult.results.channel.item.long;

    var condition_code = queryResult.results.channel.item.condition.code;

    // up to 25: display storm
    // up to 30 & 44:    display cloudy
    // up to 34: dispay  sunny
    // 35, 37-43 45-47   storm
    // 36 sunny
}
