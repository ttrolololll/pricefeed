package main

import (
	"fmt"
	"log"
	"net/http"
	"price-service/external/coindesk"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
)

// Simple http server without considerations for things like circuit breaker, timeouts and retries,
// using reference code from https://github.com/gorilla/websocket/blob/main/examples/echo/server.go

const (
	httpAddr = "localhost:8080"

	// could do json marshal to bytes, we use sprintf for simplicity for now
	// also assuming format to be json since it is not specified
	priceFeedFormat = `{"timedate":"%s","price":"%v"}`
)

var (
	upgrader = websocket.Upgrader{}
)

func main() {
	// Setup required deps
	coindesk.Init("", nil)

	http.HandleFunc("/", home)
	http.HandleFunc("/pricefeed", pricefeed)
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/pricefeed")
}

func pricefeed(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	coindeskClient := coindesk.GetClient()

	for range time.Tick(5 * time.Second) {
		resp, err := coindeskClient.CurrentPrice("BTC")
		if err != nil {
			log.Println("write:", err)
			break
		}

		// not too sure if timedate refers to time of last price update represented by resp.Time.Updated,
		// or simply the timestamp of price update reply to client,
		// using the latter as it makes more sense
		reply := fmt.Sprintf(priceFeedFormat, time.Now().UTC(), resp.BPI["USD"].RateFloat)

		err = c.WriteMessage(1, []byte(reply))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {

    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;

    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
        output.scroll(0, output.scrollHeight);
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };

    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };

});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output" style="max-height: 70vh;overflow-y: scroll;"></div>
</td></tr></table>
</body>
</html>
`))
