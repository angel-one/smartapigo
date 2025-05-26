package main

import (
	"fmt"
	SmartApi "github.com/angel-one/smartapigo"
)

func main() {

	// Create New Angel Broking Client
	ABClient := SmartApi.New("Your Client Code", "Your Password", "Your api key")

	fmt.Println("Client :- ", ABClient)

	// User Login and Generate User Session
	session, err := ABClient.GenerateSession()

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

	//Modify Order
	modifiedOrder, err := ABClient.ModifyOrder(SmartApi.ModifyOrderParams{Variety: "NORMAL", OrderID: order.OrderID, OrderType: "LIMIT", ProductType: "INTRADAY", Duration: "DAY", Price: "19400", Quantity: "1", TradingSymbol: "SBI-EQ", SymbolToken: "3045", Exchange: "NSE"})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Modified Order ID :- ", modifiedOrder)

	//Cancel Order
	cancelledOrder, err := ABClient.CancelOrder("NORMAL", modifiedOrder.OrderID)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Cancelled Order ID :- ", cancelledOrder)

	//Get Holdings
	holdings, err := ABClient.GetHoldings()

	if err != nil {
		fmt.Println(err.Error())
	} else {

		fmt.Println("Holdings :- ", holdings)
	}

	//Get Positions
	positions, err := ABClient.GetPositions()

	if err != nil {
		fmt.Println(err.Error())
	} else {

		fmt.Println("Positions :- ", positions)
	}

	//Get TradeBook
	trades, err := ABClient.GetTradeBook()

	if err != nil {
		fmt.Println(err.Error())
	} else {

		fmt.Println("All Trades :- ", trades)
	}

	//Get Last Traded Price
	ltp, err := ABClient.GetLTP(SmartApi.LTPParams{Exchange: "NSE", TradingSymbol: "SBIN-EQ", SymbolToken: "3045"})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Last Traded Price :- ", ltp)

	//Get Risk Management System
	rms, err := ABClient.GetRMS()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Risk Managemanet System :- ", rms)

	//Position Conversion
	err = ABClient.ConvertPosition(SmartApi.ConvertPositionParams{"NSE", "SBIN-EQ", "INTRADAY", "MARGIN", "BUY", 1, "DAY"})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Position Conversion Successful")
}
