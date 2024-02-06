package main

import (
	"encoding/json"
	"log"
	"net/http"
	"price-service/external/coindesk"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
)

// Simple http server without considerations for things like circuit breaker, timeouts and retries,
// using reference code from https://github.com/gorilla/websocket/blob/main/examples/echo/server.go

const (
	httpAddr                    = "localhost:8080"
	purgeClientDataOnErrTimeout = 1 * time.Minute // max duration of data bufferring while client loses connection

	// could do json marshal to bytes, we use sprintf for simplicity for now
	// also assuming format to be json since it is not specified
	priceFeedFormat = `{"timedate":"%s","price":"%v"}`
)

var (
	upgrader        = websocket.Upgrader{}
	wsClientDataMap = map[string]*wsClientData{}
)

type PriceFeed struct {
	Timedate string  `json:"timedate"`
	PriceUSD float64 `json:"price_usd,omitempty"`
	PriceEUR float64 `json:"price_eur,omitempty"`
}

type wsClientData struct {
	replyBuffer []*replyData
	expiry      *time.Time
}

type replyData struct {
	text      string
	timestamp *time.Time
}

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
	queryParams := r.URL.Query()
	clientID := queryParams.Get("client_id")
	startTimedateParam := queryParams.Get("start_timedate")
	currenciesParam := queryParams.Get("currencies")

	currencies := strings.Split(currenciesParam, ",")
	currenciesMap := map[string]struct{}{}

	if currenciesParam == "" || len(currencies) == 0 {
		currenciesMap["USD"] = struct{}{}
	}

	for _, c := range currencies {
		currenciesMap[strings.ToUpper(c)] = struct{}{}
	}

	if clientID == "" {
		log.Println("clientID not specified")
		return
	}

	// Websocket connection can be lost and unable to be reconnected by clients at any time,
	// thus use the clientID as map key to aggregate data that should be sent to the client
	// when new connection is established
	clientData, exists := wsClientDataMap[clientID]
	if !exists {
		clientData = &wsClientData{}
		wsClientDataMap[clientID] = clientData
	}

	// Upgrade HTTP connection to websocket
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	coindeskClient := coindesk.GetClient()

	// Fetch data and send to client every x seconds
	for range time.Tick(5 * time.Second) {
		// check if client data expired, purge if true,
		// also breaks the loop so that connection can be closed at server side
		if clientData.expiry != nil && time.Now().After(*clientData.expiry) {
			delete(wsClientDataMap, clientID)
			log.Println("connection expired, purge data")
			break
		}

		resp, err := coindeskClient.CurrentPrice("BTC")
		if err != nil {
			log.Println("failed to get current price:", err)
			continue
		}

		feedTimestamp := time.Now()

		feed := &PriceFeed{
			Timedate: feedTimestamp.UTC().String(),
		}

		if _, exists := currenciesMap["USD"]; exists {
			feed.PriceUSD = resp.BPI["USD"].RateFloat
		}
		if _, exists := currenciesMap["EUR"]; exists {
			feed.PriceEUR = resp.BPI["EUR"].RateFloat
		}

		replyRaw, err := json.Marshal(feed)
		if err != nil {
			log.Println("failed to marshal feed data:", err)
			continue
		}

		// not too sure if timedate refers to time of last price update represented by resp.Time.Updated,
		// or simply the timestamp of price update reply to client, here, we use the latter
		reply := string(replyRaw)

		message := ""
		replyBuffer := clientData.replyBuffer

		// if start timedate is specified, retrieve only entries on or after in reply buffer,
		// since format is not specified, will go with the assumption that query param will be in unix time format
		if startTimedateParam != "" {
			i, err := strconv.ParseInt(startTimedateParam, 10, 64)
			if err != nil {
				log.Print("parse startTimedate:", err)
				continue
			}

			startTimedate := time.Unix(i, 0)
			startIdx := 0

			for i, r := range replyBuffer {
				if startTimedate.Equal(*r.timestamp) || startTimedate.Before(*r.timestamp) {
					startIdx = i
					break
				}
			}

			replyBuffer = replyBuffer[startIdx:]
		}

		// aggregate replies while client is offline and reply in a single message
		for _, r := range replyBuffer {
			message += r.text + "\n"
		}

		message += reply

		err = c.WriteMessage(1, []byte(message))
		if err != nil {
			clientData.replyBuffer = append(clientData.replyBuffer, &replyData{text: reply, timestamp: &feedTimestamp})

			// set expiry upon first detection of client went offline
			if clientData.expiry == nil {
				exp := time.Now().Add(purgeClientDataOnErrTimeout)
				clientData.expiry = &exp
			}

			log.Println("write:", err)
			continue
		}

		// reset client data if connection restore and msg sent successfully
		clientData.replyBuffer = nil
		clientData.expiry = nil
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
        ws = new WebSocket("ws://localhost:8080/pricefeed?client_id=1&currencies=USD,EUR");
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
