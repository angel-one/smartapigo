package smartstream

import (
	"github.com/angelbroking-github/smartapigo/model"
	"log"
	"testing"
)

var client *WebSocket

func TestSmartStream(t *testing.T) {
	client = New("A586457", "00998877")
	client.callbacks.onConnected = onConnected
	client.Connect()
}

func onConnected() {
	log.Printf("connected")
	err := client.Subscribe(model.QUOTE, []model.TokenID{model.TokenID{ExchangeType: model.NSECM, Token: "26000"}})
	if err != nil {
		log.Printf("error while subscribing")
	}
}
