package smartstream

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/angelbroking-github/smartapigo/model"
	"github.com/gorilla/websocket"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type WebSocket struct {
	clientID            string
	feedToken           string
	callbacks           Callbacks
	SubsMap             map[model.SmartStreamSubsMode][]*model.TokenID
	Conn                *websocket.Conn
	url                 url.URL
	autoReconnect       bool
	reconnectMaxRetries int
	reconnectMaxDelay   time.Duration
	connectTimeout      time.Duration
	reconnectAttempt    int
	cancel              context.CancelFunc
}

//MessageHandler Handler interface for handling messages received over smartstream websocket
type Callbacks struct {
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
		SubsMap:             make(map[model.SmartStreamSubsMode][]*model.TokenID),
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

func (ws *WebSocket) Connect() error {
	return ws.ConnectWithContext(context.Background())
}

func (ws *WebSocket) ConnectWithContext(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	ws.cancel = cancel
	defer func() {
		if ws.Conn != nil {
			ws.Conn.Close()
		}
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

			if ws.reconnectAttempt > 0 {
				err = ws.resubscribe()
				if err != nil {
					return err
				}
				ws.reconnectAttempt = 0
			}

			var wg sync.WaitGroup

			// Receive stream data
			wg.Add(1)
			go ws.readMessage(ctx, &wg)

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
	if ws.callbacks.onConnected != nil {
		ws.callbacks.onConnected()
	}
}

func (ws *WebSocket) onError(err error) {
	if ws.callbacks.onError != nil {
		ws.callbacks.onError(err)
	}
}

func (ws *WebSocket) resubscribe() error {
	return nil
}

func (ws *WebSocket) onClose(code int, text string) error {
	if ws.callbacks.onClose != nil {
		ws.callbacks.onClose(code, text)
	}
	return nil
}

func (ws *WebSocket) onPing(appData string) error {
	fmt.Printf("ping received " + appData)
	return nil
}

func (ws *WebSocket) onPong(appData string) error {
	fmt.Printf("pong received " + appData)
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
	conn.SetReadDeadline(time.Now().Add(20))
	conn.SetCloseHandler(ws.onClose)
	conn.SetPingHandler(ws.onPing)
	conn.SetPongHandler(ws.onPong)
	ws.Conn = conn

	return nil

}

func (ws *WebSocket) readMessage(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			mType, msg, err := ws.Conn.ReadMessage()
			if err != nil {
				ws.onError(fmt.Errorf("Error reading data: %v", err))
				return
			}

			//Parsing binary data

			if mType == websocket.BinaryMessage {
				fmt.Printf("message received")

			} else if mType == websocket.TextMessage {
				ws.onTextMessage(msg)
			}
		}
	}
}
