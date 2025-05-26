package smartstream

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/angel-one/smartapigo/model"
	"github.com/angel-one/smartapigo/smartstream/internal/parser"
	"github.com/gorilla/websocket"
	"log"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type WebSocket struct {
	clientID            string
	feedToken           string
	callbacks           callbacks
	subsMap             map[model.SmartStreamSubsMode][]model.TokenInfo
	Conn                *websocket.Conn
	url                 url.URL
	autoReconnect       bool
	reconnectMaxRetries int
	reconnectMaxDelay   time.Duration
	connectTimeout      time.Duration
	reconnectAttempt    int
	cancel              context.CancelFunc
	lastPongTime        time.Time
	subroutineContext   context.Context
	subroutineCancel    context.CancelFunc
}

//MessageHandler Handler interface for handling messages received over smartstream websocket
type callbacks struct {
	onLTP             func(ltpInfo model.LTPInfo)
	onQuote           func(quote model.Quote)
	onSnapquote       func(quote model.SnapQuote)
	onText            func(text []byte)
	onConnected       func()
	onReconnectFailed func(reconnectAttempt int)
	onReconnect       func(attempt int, nextDelay time.Duration)
	onError           func(err error)
	onClose           func(int, string)
}

var (
	// Default ticker url.
	substreamURL = url.URL{Scheme: "ws", Host: "smartapisocket.angelone.in", Path: "/smart-stream"}
)

const (
	// Auto reconnect defaults
	// Default maximum number of reconnect attempts
	defaultReconnectMaxAttempts = 300
	// Auto reconnect min delay. Reconnect delay can't be less than this.
	reconnectMinDelay time.Duration = 5000 * time.Millisecond
	// Default auto reconnect delay to be used for auto reconnection.
	defaultReconnectMaxDelay time.Duration = 60000 * time.Millisecond
	// Connect timeout for initial server handshake.
	defaultConnectTimeout time.Duration = 7000 * time.Millisecond
	// Interval in which the connection check is performed periodically.
	connectionCheckInterval time.Duration = 10000 * time.Millisecond

	writeWait          = 5 * time.Second
	pingPeriod         = 10 * time.Second
	idleTimeout        = 15 * time.Second
	sessionCheckPeriod = 5 * time.Second

	//Headers for connection
	clientIDHeader  = "x-client-code"
	feedTokenHeader = "x-feed-token"
	clientLibHeader = "x-client-lib"
)

// New creates a new socket client  instance.
func New(clientID string, feedToken string) *WebSocket {
	ws := &WebSocket{
		clientID:            clientID,
		feedToken:           feedToken,
		url:                 substreamURL,
		autoReconnect:       true,
		reconnectMaxDelay:   defaultReconnectMaxDelay,
		reconnectMaxRetries: defaultReconnectMaxAttempts,
		connectTimeout:      defaultConnectTimeout,
		subsMap:             make(map[model.SmartStreamSubsMode][]model.TokenInfo),
	}

	return ws
}

// SetRootURL sets ticker root url.
func (ws *WebSocket) SetRootURL(u url.URL) {
	ws.url = u
}

// SetAccessToken set access token.
func (ws *WebSocket) SetFeedToken(feedToken string) {
	ws.feedToken = feedToken
}

// SetConnectTimeout sets default timeout for initial connect handshake
func (ws *WebSocket) SetConnectTimeout(val time.Duration) {
	ws.connectTimeout = val
}

// SetAutoReconnect enable/disable auto reconnect.
func (ws *WebSocket) SetAutoReconnect(val bool) {
	ws.autoReconnect = val
}

// SetReconnectMaxDelay sets maximum auto reconnect delay.
func (ws *WebSocket) SetReconnectMaxDelay(val time.Duration) error {
	if val > reconnectMinDelay {
		return fmt.Errorf("ReconnectMaxDelay can't be less than %fms", reconnectMinDelay.Seconds()*1000)
	}

	ws.reconnectMaxDelay = val
	return nil
}

// SetReconnectMaxRetries sets maximum reconnect attempts.
func (ws *WebSocket) SetReconnectMaxRetries(val int) {
	ws.reconnectMaxRetries = val
}

func (ws *WebSocket) SetOnConnected(fn func()) {
	if fn != nil {
		ws.callbacks.onConnected = fn
	}
}

func (ws *WebSocket) SetOnSnapquote(fn func(model.SnapQuote)) {
	if fn != nil {
		ws.callbacks.onSnapquote = fn
	}
}

func (ws *WebSocket) SetOnLTP(fn func(info model.LTPInfo)) {
	if fn != nil {
		ws.callbacks.onLTP = fn
	}
}

func (ws *WebSocket) SetOnQuote(fn func(quote model.Quote)) {
	if fn != nil {
		ws.callbacks.onQuote = fn
	}
}

func (ws *WebSocket) SetOnError(fn func(err error)) {
	if fn != nil {
		ws.callbacks.onError = fn
	}
}

func (ws *WebSocket) SetOnReconnect(fn func(attempt int, nextDelay time.Duration)) {
	if fn != nil {
		ws.callbacks.onReconnect = fn
	}
}

func (ws *WebSocket) SetOnReconnectFailed(fn func(attempt int)) {
	if fn != nil {
		ws.callbacks.onReconnectFailed = fn
	}
}

func (ws *WebSocket) SetOnClose(fn func(int, string)) {
	if fn != nil {
		ws.callbacks.onClose = fn
	}
}

func (ws *WebSocket) Connect() error {
	return ws.ConnectWithContext(context.Background())
}

func (ws *WebSocket) ConnectWithContext(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	ws.cancel = cancel
	defer func() {
		ws.Stop()
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if ws.reconnectAttempt > ws.reconnectMaxRetries {
				ws.onReconnectFailed(ws.reconnectAttempt)
			}
			if ws.reconnectAttempt > 0 {
				nextDelay := time.Duration(math.Pow(2, float64(ws.reconnectAttempt))) * time.Second
				if nextDelay > ws.reconnectMaxDelay || nextDelay <= 0 {
					nextDelay = ws.reconnectMaxDelay
				}

				ws.onReconnect(ws.reconnectAttempt, nextDelay)
				log.Printf("attempting reconnect in %f seconds", nextDelay.Seconds())
				time.Sleep(nextDelay)

				if ws.Conn != nil { // Closing previous connection
					ws.Conn.Close()
				}
			}

			err := ws.createConnection()

			if err != nil {
				ws.onError(err)
				if ws.autoReconnect {
					ws.reconnectAttempt++
					continue
				}
				return err
			}
			ws.onConnected()
			ws.subroutineContext, ws.subroutineCancel = context.WithCancel(context.Background())
			go ws.startPing()

			var wg sync.WaitGroup

			// Receive stream data
			wg.Add(1)
			go ws.readMessage(&wg)

			wg.Add(1)
			go ws.checkIdleConnection(&wg)

			wg.Wait()

		}
	}
}

func (ws *WebSocket) onReconnectFailed(reconnectAttempt int) {
	if ws.callbacks.onReconnectFailed != nil {
		ws.callbacks.onReconnectFailed(reconnectAttempt)
	}

}

func (ws *WebSocket) onReconnect(attempt int, delay time.Duration) {
	if ws.callbacks.onReconnect != nil {
		ws.callbacks.onReconnect(attempt, delay)
	}
}

func (ws *WebSocket) onConnected() {

	if ws.reconnectAttempt > 0 {
		err := ws.resubscribe()
		if err != nil {
			return
		}
		ws.reconnectAttempt = 0
	} else {
		if ws.callbacks.onConnected != nil {
			ws.callbacks.onConnected()
		}
	}
}

func (ws *WebSocket) onError(err error) {
	if ws.callbacks.onError != nil {
		ws.callbacks.onError(err)
	}
}

func (ws *WebSocket) resubscribe() (err error) {
	for k, v := range ws.subsMap {
		err = ws.subscribeToTokens(k, v)
		if err != nil {
			return
		}
	}
	return
}

func (ws *WebSocket) Subscribe(mode model.SmartStreamSubsMode, tokenIds []model.TokenInfo) error {
	err := ws.subscribeToTokens(mode, tokenIds)
	if err == nil {
		if _, ok := ws.subsMap[mode]; !ok {
			ws.subsMap[mode] = make([]model.TokenInfo, 0)
		}
		ws.subsMap[mode] = append(ws.subsMap[mode], tokenIds...)
	}
	return err
}

func (ws *WebSocket) subscribeToTokens(mode model.SmartStreamSubsMode, tokenIds []model.TokenInfo) error {
	request, err := ws.createSubsRequest(mode, tokenIds)
	if err != nil {
		return err
	}
	err = ws.Conn.WriteMessage(websocket.TextMessage, request)
	return err
}

func (ws *WebSocket) onClose(code int, text string) error {
	fmt.Printf("connection closed ")
	if ws.callbacks.onClose != nil {
		ws.callbacks.onClose(code, text)
	}
	return nil
}

func (ws *WebSocket) Stop() {
	ws.closeRoutines()
	if ws.cancel != nil {
		ws.cancel()
	}
}

func (ws *WebSocket) closeRoutines() {
	if ws.subroutineCancel != nil {
		ws.subroutineCancel()
		if ws.Conn != nil {
			ws.Conn.Close()
		}
	}
}

func (ws *WebSocket) onPong(appData string) error {
	ws.lastPongTime = time.Now()
	return nil
}

func (ws *WebSocket) onTextMessage(text []byte) {
	if ws.callbacks.onText != nil {
		ws.callbacks.onText(text)
	}
}

func (ws *WebSocket) createConnection() error {
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = ws.connectTimeout
	dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	headers := http.Header{}
	headers.Add(clientIDHeader, ws.clientID)
	headers.Add(feedTokenHeader, ws.feedToken)
	headers.Add(clientLibHeader, "GOLANG")

	conn, _, err := dialer.Dial(ws.url.String(), headers)
	if err != nil {
		return err
	}
	conn.SetCloseHandler(ws.onClose)
	conn.SetPongHandler(ws.onPong)
	ws.Conn = conn

	return nil

}

func (ws *WebSocket) readMessage(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ws.subroutineContext.Done():
			return
		default:
			mType, msg, err := ws.Conn.ReadMessage()
			if err != nil {
				ws.onError(fmt.Errorf("Error reading data: %v", err))
				return
			}

			//Parsing binary data
			if mType == websocket.BinaryMessage {
				mode := model.SmartStreamSubsMode(msg[0])

				switch mode {
				case model.LTP:
					ltp := parser.ParseLTP(msg)
					ws.callbacks.onLTP(ltp)
				case model.QUOTE:
					quote := parser.ParseQuote(msg)
					ws.callbacks.onQuote(quote)
				case model.SNAPQUOTE:
					snapquote := parser.ParseSnapquote(msg)
					ws.callbacks.onSnapquote(snapquote)
				default:
					log.Printf("Message mode not  recognized")
				}

			} else if mType == websocket.TextMessage {
				ws.onTextMessage(msg)
			}
		}
	}
}

func (ws *WebSocket) startPing() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ws.subroutineContext.Done():
			return
		case <-ticker.C:
			ws.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.Conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				return
			}
		}
	}
}

func (ws *WebSocket) checkIdleConnection(wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(pingPeriod * 2)
	ticker := time.NewTicker(sessionCheckPeriod)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ticker.C:
			if time.Since(ws.lastPongTime).Seconds() > idleTimeout.Seconds() {
				log.Printf("ping not received. Reconnecting ...")
				ws.closeRoutines()
				ws.reconnectAttempt++
				return
			}
		}
	}
}

func (ws *WebSocket) createSubsRequest(mode model.SmartStreamSubsMode, tokenIds []model.TokenInfo) ([]byte, error) {

	exchangeTokenMap := make(map[model.ExchangeType][]string)
	for _, val := range tokenIds {
		if _, ok := exchangeTokenMap[val.ExchangeType]; !ok {
			exchangeTokenMap[val.ExchangeType] = make([]string, 0)
		}
		exchangeTokenMap[val.ExchangeType] = append(exchangeTokenMap[val.ExchangeType], val.Token)
	}

	tokenList := make([]model.SubscriptionTokens, 0)
	for k, v := range exchangeTokenMap {
		subscriptionTokens := model.SubscriptionTokens{ExchangeType: k, Tokens: v}
		tokenList = append(tokenList, subscriptionTokens)
	}
	params := model.SubscriptionParam{Mode: mode, TokenList: tokenList}

	subscriptionRequest := model.SubscriptionRequest{}
	subscriptionRequest.Action = 1
	subscriptionRequest.Params = params
	subscriptionRequest.CorrelationID = "abc"

	return json.Marshal(subscriptionRequest)

}
