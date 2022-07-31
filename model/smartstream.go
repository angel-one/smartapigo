package model

type ExchangeType int
type SmartStreamAction int8
type SmartStreamSubsMode int8

const BYTES int = 20

const (
	NSECM ExchangeType = 1
	NSEFO ExchangeType = 2
	BSECM ExchangeType = 3
	BSEFO ExchangeType = 4
	MCXFO ExchangeType = 5
	NCXFO ExchangeType = 7
	CDEFO ExchangeType = 13
)

const (
	SUBS   SmartStreamAction = 1
	UNSUBS SmartStreamAction = 0
)

const (
	LTP       SmartStreamSubsMode = 1
	QUOTE     SmartStreamSubsMode = 2
	SNAPQUOTE SmartStreamSubsMode = 3
)

type TokenID struct {
	ExchangeType ExchangeType
	Token        string
}

type SmartApiBBSInfo struct {
	SiBbBuySellFlag  int16
	lQuantity        int64
	lPrice           int64
	SiNumberOfOrders int16
}

type LTPInfo struct {
	TokenID                     TokenID
	SequenceNumber              int64
	ExchangeFeedTimeEpochMillis int64
	LastTradedPrice             int64
}

type Quote struct {
	TokenID                     TokenID
	SequenceNumber              int64
	ExchangeFeedTimeEpochMillis int64
	LastTradedPrice             int64
	LastTradedQty               int64
	AvgTradedPrice              int64
	VolumeTradedToday           int64
	TotalBuyQty                 float64
	TotalSellQty                float64
	OpenPrice                   int64
	HighPrice                   int64
	LowPrice                    int64
	ClosePrice                  int64
}

type SnapQuote struct {
	TokenID                     TokenID
	SequenceNumber              int64
	ExchangeFeedTimeEpochMillis int64
	LastTradedPrice             int64
	LastTradedQty               int64
	AvgTradedPrice              int64
	VolumeTradedToday           int64
	TotalBuyQty                 float64
	TotalSellQty                float64
	OpenPrice                   int64
	HighPrice                   int64
	LowPrice                    int64
	ClosePrice                  int64
	LastTradedTimestamp         int64
	OpenInterest                int64
	OpenInterestChangePerc      float64
	BestFiveBuy                 []*SmartApiBBSInfo
	BestFiveSell                []*SmartApiBBSInfo
	UpperCircuit                int64
	LowerCircuit                int64
	YearlyHighPrice             int64
	YearlyLowPrice              int64
}

type SubscriptionRequest struct {
	CorrelationID string            `json:"correlationID"`
	Action        int8              `json:"action"`
	Params        SubscriptionParam `json:"params"`
}

type SubscriptionParam struct {
	Mode      SmartStreamSubsMode  `json:"mode"`
	TokenList []SubscriptionTokens `json:"tokenList"`
}

type SubscriptionTokens struct {
	ExchangeType ExchangeType `json:"exchangeType"`
	Tokens       []string     `json:"tokens"`
}
