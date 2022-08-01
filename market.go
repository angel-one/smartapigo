package smartapigo

import (
	"net/http"
)

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

// GetLTP gets Last Traded Price.
func (c *Client) GetLTP(ltpParams LTPParams) (LTPResponse, error) {
	var ltp LTPResponse
	params := structToMap(ltpParams, "json")
	err := c.doEnvelope(http.MethodPost, URILTP, params, nil, &ltp, true)
	return ltp, err
}
