window.addEventListener("load", init);

function init() {
    xhr = new XMLHttpRequest();
    xhr.addEventListener("load", reqHandler);
    xhr.open("GET", "getweatherinfo");
    xhr.send();
}

function reqHandler(evt) {
    console.log(JSON.parse(evt.responseText));
}
