# BTC Price Feed

## Part 1

### Throught Process & Approach

Took some time to revise on the various server push mechanisms. Two that immediately came to mind are websockets and server-sent events. Websocket is used to implement the requirements given the unfamiliarity with SSE and time available.

The websocket endpoint is implemented using the Gorilla Websocket library, which comes with a starter file that includes a frontend page that could be used for testing. The starter file was used in the exercise.

The implementation uses time intervals of 5 seconds, and within each interval, API call is made to Coindesk to retrieve the data. The data is then parsed and pushed to the client.

The Coindesk client uses the standard HTTP client and resiliency safeguards like retries and circuit-breaker are omitted.

Tests are omitted at this point.

### Usage
1. Startup server `go run main.go`
1. Visit `http://localhost:8080/home`, subscribe to price feed by clicking `Open` and observe BTC price every 5 seconds

## Part 2

### Throught Process & Approach

The major challenge for part 2 is that for websocket, once connection is closed by client, the client cannot re-establish the connection. I.e UserA opens a connection in browser, user refreshed browser and connection is lost, UserA re-opens a connection, the connection will be a new connection without context of the previous connection.

A possible way to resolve this is to have a concept of a client ID. A seperate client ID endpoint could be created for user to obtain a client ID prior to subscribing. Alternatively, the same price feed endpoint could generate one if not present, and reply back to the client. With the client ID, if a connection is lost and a new connection is established, information can still be retrieved via the client ID in the server. The implementation assumes client ID is known and will be provided when calling the price feed endpoint.

The implementation makes use of the server push step, where if there is an error (eg. client closed connection), it stores the data for the client ID. Upon an error, it also sets an expiry for when the connection should be killed by the server. I.e if a client went offline for too long, stop aggregating the data and purge any stored data. If `start_timedate` is provided as a query param, it will be used to only display availabled data at or after `start_timedate`. `start_timedate` is assumed to be in UNIX seconds format for the purpose of the exercise.

### Usage
1. Same as part 1
1. Refresh browser to simulate browser lost the connection, wait for ~10 seconds and reopen connection
1. On first reply, client should receive the data including those missed during the ~10 seconds client offline period

## Part 3

### Throught Process & Approach

The implementation introduces a new query param expecting the value to be a comma-seperated list of currencies the client wants to receive the price in.

### Usage
1. Modify currencies param in `ws://localhost:8080/pricefeed?client_id=1&currencies=USD,EUR` in `main.go` and execute