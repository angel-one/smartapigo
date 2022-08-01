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

type TokenInfo struct {
	ExchangeType ExchangeType
	Token        string
}

type SmartApiBBSInfo struct {
	Flag           uint16
	Quantity       uint64
	Price          uint64
	NumberOfOrders uint16
}

type LTPInfo struct {
	TokenInfo                   TokenInfo
	SequenceNumber              uint64
	ExchangeFeedTimeEpochMillis uint64
	LastTradedPrice             uint64
}

type Quote struct {
	TokenInfo                   TokenInfo
	SequenceNumber              uint64
	ExchangeFeedTimeEpochMillis uint64
	LastTradedPrice             uint64
	LastTradedQty               uint64
	AvgTradedPrice              uint64
	VolumeTradedToday           uint64
	TotalBuyQty                 float64
	TotalSellQty                float64
	OpenPrice                   uint64
	HighPrice                   uint64
	LowPrice                    uint64
	ClosePrice                  uint64
}

type SnapQuote struct {
	TokenInfo                   TokenInfo
	SequenceNumber              uint64
	ExchangeFeedTimeEpochMillis uint64
	LastTradedPrice             uint64
	LastTradedQty               uint64
	AvgTradedPrice              uint64
	VolumeTradedToday           uint64
	TotalBuyQty                 float64
	TotalSellQty                float64
	OpenPrice                   uint64
	HighPrice                   uint64
	LowPrice                    uint64
	ClosePrice                  uint64
	LastTradedTimestamp         uint64
	OpenInterest                uint64
	OpenInterestChangePerc      float64
	BestFiveBuy                 []SmartApiBBSInfo
	BestFiveSell                []SmartApiBBSInfo
	UpperCircuit                uint64
	LowerCircuit                uint64
	YearlyHighPrice             uint64
	YearlyLowPrice              uint64
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
