# The Smart API Go client

The official Go client for communicating with the Angel Broking Smart APIs.

SmartAPI is a set of REST-like APIs that expose many capabilities required to build a complete investment and trading platform. Execute orders in real time, manage user portfolio, stream live market data (WebSockets), and more, with the simple HTTP API collection.


## Installation
```
go get github.com/angel-one/smartapigo
```
## API usage
```golang
package main

import (
	"fmt"
	SmartApi "github.com/angel-one/smartapigo"
)

func main() {

	// Create New Angel Broking Client
	ABClient := SmartApi.New("ClientCode", "Password","API Key")

	fmt.Println("Client :- ",ABClient)

	// User Login and Generate User Session
	session, err := ABClient.GenerateSession("totp here")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//Renew User Tokens using refresh token
	session.UserSessionTokens, err = ABClient.RenewAccessToken(session.RefreshToken)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("User Session Tokens :- ", session.UserSessionTokens)

	//Get User Profile
	session.UserProfile, err = ABClient.GetUserProfile()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("User Profile :- ", session.UserProfile)
	fmt.Println("User Session Object :- ", session)

	//Place Order
	order, err := ABClient.PlaceOrder(SmartApi.OrderParams{Variety: "NORMAL", TradingSymbol: "SBIN-EQ", SymbolToken: "3045", TransactionType: "BUY", Exchange: "NSE", OrderType: "LIMIT", ProductType: "INTRADAY", Duration: "DAY", Price: "19500", SquareOff: "0", StopLoss: "0", Quantity: "1"})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Placed Order ID and Script :- ", order)
}
```
## Websocket Data Streaming
```golang
package main

import (
	"fmt"
	SmartApi "github.com/angel-one/smartapigo"
	"github.com/angel-one/smartapigo/websocket"
	"time"
)

var socketClient *websocket.SocketClient

// Triggered when any error is raised
func onError(err error) {
	fmt.Println("Error: ", err)
}

// Triggered when websocket connection is closed
func onClose(code int, reason string) {
	fmt.Println("Close: ", code, reason)
}

// Triggered when connection is established and ready to send and accept data
func onConnect() {
	fmt.Println("Connected")
	err := socketClient.Subscribe()
	if err != nil {
		fmt.Println("err: ", err)
	}
}

// Triggered when a message is received
func onMessage(message []map[string]interface{})  {
	fmt.Printf("Message Received :- %v\n",message)
}

// Triggered when reconnection is attempted which is enabled by default
func onReconnect(attempt int, delay time.Duration) {
	fmt.Printf("Reconnect attempt %d in %fs\n", attempt, delay.Seconds())
}

// Triggered when maximum number of reconnect attempt is made and the program is terminated
func onNoReconnect(attempt int) {
	fmt.Printf("Maximum no of reconnect attempt reached: %d\n", attempt)
}

func main() {

	// Create New Angel Broking Client
	ABClient := SmartApi.New("ClientCode", "Password","API Key")

	// User Login and Generate User Session
	session, err := ABClient.GenerateSession()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//Get User Profile
	session.UserProfile, err = ABClient.GetUserProfile()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// New Websocket Client
	socketClient = websocket.New(session.ClientCode,session.FeedToken,"nse_cm|17963&nse_cm|3499&nse_cm|11536&nse_cm|21808&nse_cm|317")

	// Assign callbacks
	socketClient.OnError(onError)
	socketClient.OnClose(onClose)
	socketClient.OnMessage(onMessage)
	socketClient.OnConnect(onConnect)
	socketClient.OnReconnect(onReconnect)
	socketClient.OnNoReconnect(onNoReconnect)

	// Start Consuming Data
	socketClient.Serve()

}
```

## Examples
Check example folder for more examples.

You can run the following after updating the Credentials in the examples:
```
go run example/example.go
```
For websocket example
```
go run example/websocket/example.go
```

## Run unit tests

```
go test -v
```
