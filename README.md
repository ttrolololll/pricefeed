# BTC Price Feed

## Part 1

### Thought Process & Approach

Took some time to revise on the various server push mechanisms. Two that immediately came to mind are websockets and server-sent events. Websocket is used to implement the requirements given the unfamiliarity with SSE and time available.

The websocket endpoint is implemented using the Gorilla Websocket library, which comes with a starter file that includes a frontend page that could be used for testing. The starter file was used in the exercise.

The implementation uses time intervals of 5 seconds, and within each interval, API call is made to Coindesk to retrieve the data. The data is then parsed and pushed to the client.

The Coindesk client uses the standard HTTP client and resiliency safeguards like retries and circuit-breaker are omitted.

Tests are omitted at this point.

### Usage
1. Startup server `go run main.go`
1. Visit `http://localhost:8080/home`, subscribe to price feed by clicking `Open` and observe BTC price every 5 seconds

## Part 2

### Thought Process & Approach

The major challenge for part 2 is that for websocket, once connection is closed by client, the client cannot re-establish the connection. I.e UserA opens a connection in browser, user refreshed browser and connection is lost, UserA re-opens a connection, the connection will be a new connection without context of the previous connection.

A possible way to resolve this is to have a concept of a client ID. A seperate client ID endpoint could be created for user to obtain a client ID prior to subscribing. Alternatively, the same price feed endpoint could generate one if not present, and reply back to the client. With the client ID, if a connection is lost and a new connection is established, information can still be retrieved via the client ID in the server. The implementation assumes client ID is known and will be provided when calling the price feed endpoint.

The implementation makes use of the server push step, where if there is an error (eg. client closed connection), it stores the data for the client ID. Upon an error, it also sets an expiry for when the connection should be killed by the server. I.e if a client went offline for too long, stop aggregating the data and purge any stored data. If `start_timedate` is provided as a query param, it will be used to only display availabled data at or after `start_timedate`. `start_timedate` is assumed to be in UNIX seconds format for the purpose of the exercise.

### Usage
1. Same as part 1
1. Refresh browser to simulate browser lost the connection, wait for ~10 seconds and reopen connection
1. On first reply, client should receive the data including those missed during the ~10 seconds client offline period

## Part 3

### Thought Process & Approach

The implementation introduces a new query param expecting the value to be a comma-seperated list of currencies the client wants to receive the price in.

### Usage
1. Modify currencies param in `ws://localhost:8080/pricefeed?client_id=1&currencies=USD,EUR` in `main.go` and execute

## Part 4

### Thought Process & Approach

- To do horizontal scaling, we would deploy multiple instances behind a load-balancer
- The current way of aggregating data will not work as the data is store in-memory within a single server instance. If client reconnects to a different server instance, no aggregated data will be available.
- To resolve, the service will need a storage accessible by all server instances. Since the data is short-lived and periodically purged, Redis would be a good choice for fast retrivals and TTLs.

Code could be refactored for better maintainability and testibility. Some unit tests are written here.

## Part 5

### Thought Process & Approach

#### Architectural

The current implementation starts up a process with 5-second work intervals for ALL clients. There will be a lot of duplicate processings, eg. UserA and UserB both made connections at T1. Both get same update at T+5s, T+10s, etc. The data fetch from Coindesk and processing ideally should just happen once and pushed to both UserA and UserB.

This suggests that the data fetching and processing could be a scheduled system task instead and the results persisted into storage. Storing into DB allows for the building of price history, and allows for better capabilities to aggregate and send missed updates to clients.

The updates to the client also do not need to be a fixed `x` second interval. The process can be optmised by only sending updates to clients when the newly fetched price differs from last known price. This can be acheived by adopting a more event-driven approach with pub/sub:

- Scheduled system task executes and fetches latest price from source
- Persist into storage, get `lastKnownPrice` from Redis
- If differs from price from fetched data, update `lastKnownPrice` cache value, publish a new message
- Each active websocket connection within the server will be initialised as a subscriber to the pub/sub topic 
- When messages are received by the subscribers, server push updates to clients

#### Resiliency

- The API call to the source to retrive price data is prone to failure, a retry policy should be defined
- Currently there is only one single source, in an event Coindesk went down, the service will be down as well with incomplete record
- Have more than one source 
- Since the endpoint is available to public, it will be proned to DDoS. As such, a rate-limiting mechanism could be used on the endpoint

#### Availability

- To handle legitimate traffic spikes, the load-balancer will need to be able to perform auto-scaling
- Rate limit the APIs to prevent exceeding