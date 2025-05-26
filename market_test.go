package smartapigo

import (
	"testing"
)

func (ts *TestSuite) TestGetLTP(t *testing.T) {
	t.Parallel()
	params := LTPParams{
		Exchange:      "NSE",
		TradingSymbol: "SBIN-EQ",
		SymbolToken:   "3045",
	}
	ltp, err := ts.TestConnect.GetLTP(params)
	if err != nil {
		t.Errorf("Error while fetching LTP. %v", err)
	}

	if ltp.Exchange == "" {
		t.Errorf("Error while exchange in LTP. %v", err)
	}

}

func (ts *TestSuite) TestLTPForTokens(t *testing.T) {
	t.Parallel()
	params := LTPForTokens{
		Mode: "LTP",
	}
	params.ExchangeTokens.Nse = []string{"3045", "2885"}
	ltp, err := ts.TestConnect.GetLTPForTokens(params)
	if err != nil {
		t.Errorf("Error while fetching LTP. %v", err)
	}

	if ltp.Status == false {
		t.Errorf("Error while exchange in LTP for tokens. %v", err)
	}

}
