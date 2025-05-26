package parser

import (
	"encoding/binary"
	"github.com/angel-one/smartapigo/model"
	"math"
)

func ParseLTP(msg []byte) model.LTPInfo {
	sequenceNumer := binary.LittleEndian.Uint64(msg[27:35])
	exchangeTimestamp := binary.LittleEndian.Uint64(msg[35:43])
	ltp := binary.LittleEndian.Uint64(msg[43:51])

	ltpInfo := model.LTPInfo{TokenInfo: getToken(msg), SequenceNumber: sequenceNumer,
		ExchangeFeedTimeEpochMillis: exchangeTimestamp, LastTradedPrice: ltp}

	return ltpInfo
}

func ParseQuote(msg []byte) model.Quote {
	quote := model.Quote{}
	ltpInfo := ParseLTP(msg)
	quote.SequenceNumber = ltpInfo.SequenceNumber
	quote.TokenInfo = ltpInfo.TokenInfo
	quote.ExchangeFeedTimeEpochMillis = ltpInfo.ExchangeFeedTimeEpochMillis
	quote.LastTradedPrice = ltpInfo.LastTradedPrice
	quote.LastTradedQty = binary.LittleEndian.Uint64(msg[51:59])
	quote.AvgTradedPrice = binary.LittleEndian.Uint64(msg[59:67])
	quote.VolumeTradedToday = binary.LittleEndian.Uint64(msg[67:75])
	quote.TotalBuyQty = math.Float64frombits(binary.LittleEndian.Uint64(msg[75:83]))
	quote.TotalSellQty = math.Float64frombits(binary.LittleEndian.Uint64(msg[83:91]))
	quote.OpenPrice = binary.LittleEndian.Uint64(msg[91:99])
	quote.HighPrice = binary.LittleEndian.Uint64(msg[99:107])
	quote.LowPrice = binary.LittleEndian.Uint64(msg[107:115])
	quote.ClosePrice = binary.LittleEndian.Uint64(msg[115:123])
	return quote
}

func ParseSnapquote(msg []byte) model.SnapQuote {
	snapquote := model.SnapQuote{}
	quote := ParseQuote(msg)
	snapquote.SequenceNumber = quote.SequenceNumber
	snapquote.TokenInfo = quote.TokenInfo
	snapquote.ExchangeFeedTimeEpochMillis = quote.ExchangeFeedTimeEpochMillis
	snapquote.LastTradedPrice = quote.LastTradedPrice
	snapquote.LastTradedQty = quote.LastTradedQty
	snapquote.AvgTradedPrice = quote.AvgTradedPrice
	snapquote.VolumeTradedToday = quote.VolumeTradedToday
	snapquote.TotalBuyQty = quote.TotalBuyQty
	snapquote.TotalSellQty = quote.TotalSellQty
	snapquote.OpenPrice = quote.OpenPrice
	snapquote.HighPrice = quote.HighPrice
	snapquote.LowPrice = quote.LowPrice
	snapquote.ClosePrice = quote.ClosePrice

	snapquote.LastTradedTimestamp = binary.LittleEndian.Uint64(msg[123:131])
	snapquote.OpenInterest = binary.LittleEndian.Uint64(msg[131:139])
	snapquote.OpenInterestChangePerc = math.Float64frombits(binary.LittleEndian.Uint64(msg[139:147]))

	snapquote.BestFiveBuy, snapquote.BestFiveSell = getBestBuySellData(msg[147:347])

	snapquote.UpperCircuit = binary.LittleEndian.Uint64(msg[347:355])
	snapquote.LowerCircuit = binary.LittleEndian.Uint64(msg[355:363])
	snapquote.YearlyHighPrice = binary.LittleEndian.Uint64(msg[363:371])
	snapquote.YearlyLowPrice = binary.LittleEndian.Uint64(msg[371:379])

	return snapquote
}

func getBestBuySellData(msg []byte) (bestFiveBuy []model.SmartApiBBSInfo, bestFiveSell []model.SmartApiBBSInfo) {
	bestFiveBuy = make([]model.SmartApiBBSInfo, 0)
	bestFiveSell = make([]model.SmartApiBBSInfo, 0)
	for i := 0; i < 200; i = i + 20 {
		info := model.SmartApiBBSInfo{}
		info.Flag = binary.LittleEndian.Uint16(msg[i : i+2])
		info.Quantity = binary.LittleEndian.Uint64(msg[i+2 : i+10])
		info.Price = binary.LittleEndian.Uint64(msg[i+10 : i+18])
		info.NumberOfOrders = binary.LittleEndian.Uint16(msg[i+18 : i+20])
		if info.Flag == 1 {
			bestFiveBuy = append(bestFiveBuy, info)
		} else {
			bestFiveSell = append(bestFiveSell, info)
		}

	}

	return
}

func getToken(msg []byte) model.TokenInfo {
	exchangeType := model.ExchangeType(msg[1])
	tokenEnd := 0

	for i := 2; i < 27; i++ {
		tokenEnd++
		if int(msg[i]) == 0 {
			break
		}
	}
	return model.TokenInfo{ExchangeType: exchangeType, Token: string(msg[2 : tokenEnd+1])}
}
