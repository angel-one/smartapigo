package smartapigo

import (
	"net/http"
)

// Holding is an individual holdings response.
type Holding struct {
	Tradingsymbol      string  `json:"tradingsymbol"`
	Exchange           string  `json:"exchange"`
	ISIN               string  `json:"isin"`
	T1Quantity         int64   `json:"t1quantity"`
	RealisedQuantity   int64   `json:"realisedquantity"`
	Quantity           int64   `json:"quantity"`
	AveragePrice       float64 `json:"averageprice"`
	AuthorisedQuantity int64   `json:"authorisedquantity"`
	ProfitAndLoss      float64 `json:"profitandloss"`
	Product            string  `json:"product"`
	CollateralQuantity int64   `json:"collateralquantity"`
	CollateralType     string  `json:"collateraltype"`
	Haircut            float64 `json:"haircut"`
	Ltp                float64 `json:"ltp"`
	Symboltoken        string  `json:"symboltoken"`
	Close              float64 `json:"close"`
	Pnlpercentage      float64 `json:"pnlpercentage"`
}

// Holdings is a list of holdings
type Holdings []Holding

// GetHoldings gets a list of holdings.
func (c *Client) GetHoldings() (Holdings, error) {
	var holdings Holdings
	err := c.doEnvelope(http.MethodGet, URIGetHoldings, nil, nil, &holdings, true)
	return holdings, err
}
