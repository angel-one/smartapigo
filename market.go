package smartapigo

import "net/http"

// LTPResponse represents LTP API Response.
type LTPResponse struct {
	Exchange      string  `json:"exchange"`
	TradingSymbol string  `json:"tradingsymbol"`
	SymbolToken   string  `json:"symboltoken"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	Ltp           float64 `json:"ltp"`
}

// LTPParams represents parameters for getting LTP.
type LTPParams struct {
	Exchange      string `json:"exchange"`
	TradingSymbol string `json:"tradingsymbol"`
	SymbolToken   string `json:"symboltoken"`
}

type LTPForTokens struct {
	Mode           string `json:"mode"`
	ExchangeTokens struct {
		Nse []string `json:"NSE"`
		BSE []string `json:"BSE"`
		NFO []string `json:"NFO"`
		BFO []string `json:"BFO"`
		CDS []string `json:"CDS"`
		MCX []string `json:"MCX"`
	} `json:"exchangeTokens"`
}

type LTPMultipleResponse struct {
	Status    bool   `json:"status"`
	Message   string `json:"message"`
	Errorcode string `json:"errorcode"`
	Data      struct {
		Fetched []struct {
			Exchange      string  `json:"exchange"`
			TradingSymbol string  `json:"tradingSymbol"`
			Token         string  `json:"symbolToken"`
			Ltp           float64 `json:"ltp"`
		} `json:"fetched"`
		Unfetched []struct {
			Exchange  string `json:"exchange"`
			Token     string `json:"symbolToken"`
			Message   string `json:"message"`
			ErrorCode string `json:"errorCode"`
		} `json:"unfetched"`
	} `json:"data"`
}

// GetLTP gets Last Traded Price.
func (c *Client) GetLTP(ltpParams LTPParams) (LTPResponse, error) {
	var ltp LTPResponse
	params := structToMap(ltpParams, "json")
	err := c.doEnvelope(http.MethodPost, URILTP, params, nil, &ltp, true)
	return ltp, err
}

// GetLTPForTokens gets Last Traded Price for all tokens passed.
func (c *Client) GetLTPForTokens(ltpParams LTPForTokens) (LTPMultipleResponse, error) {
	var ltp LTPMultipleResponse
	params := structToMap(ltpParams, "json")
	err := c.doEnvelope(http.MethodPost, URILTPQuote, params, nil, &ltp, true)
	return ltp, err
}
