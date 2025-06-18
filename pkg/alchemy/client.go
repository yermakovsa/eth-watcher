package alchemy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/coder/websocket"
)

const (
	alchemyWSURL = "wss://eth-mainnet.g.alchemy.com/v2/"
)

type WSConn interface {
	Read(ctx context.Context) (websocket.MessageType, []byte, error)
	Write(ctx context.Context, typ websocket.MessageType, p []byte) error
	Close(code websocket.StatusCode, reason string) (err error)
}

// AlchemyClient represents the abstraction over Alchemy WebSocket
type AlchemyClient struct {
	conn       WSConn
	apiKey     string
	ctx        context.Context
	cancel     context.CancelFunc
	writeMu    sync.Mutex
	msgChan    chan MinedTxEvent
	errorChan  chan error
	closedChan chan struct{}
}

func NewAlchemyClient(apiKey string) (*AlchemyClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	url := alchemyWSURL + apiKey

	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to Alchemy WS: %w", err)
	}

	client := &AlchemyClient{
		conn:       conn,
		apiKey:     apiKey,
		ctx:        ctx,
		cancel:     cancel,
		msgChan:    make(chan MinedTxEvent, 100),
		errorChan:  make(chan error, 1),
		closedChan: make(chan struct{}),
	}

	return client, nil
}

func NewAlchemyClientWithConn(ctx context.Context, conn WSConn) *AlchemyClient {
	return &AlchemyClient{
		conn:       conn,
		ctx:        ctx,
		cancel:     func() {}, // optionally allow cancellation
		msgChan:    make(chan MinedTxEvent, 100),
		errorChan:  make(chan error, 1),
		closedChan: make(chan struct{}),
	}
}

func (a *AlchemyClient) Close() error {
	a.cancel()
	close(a.closedChan)
	return a.conn.Close(websocket.StatusNormalClosure, "client closed")
}

func (a *AlchemyClient) SubscribeMinedTransactions(opts MinedTxOptions) (<-chan MinedTxEvent, error) {
	subReq := SubscriptionRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_subscribe",
		Params:  []any{"alchemy_minedTransactions", opts},
	}

	data, err := json.Marshal(subReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subscription request: %w", err)
	}

	a.writeMu.Lock()
	err = a.conn.Write(a.ctx, websocket.MessageText, data)
	a.writeMu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	// Start listening for messages
	go a.readLoop()
	return a.msgChan, nil
}

func (a *AlchemyClient) readLoop() {
	for {
		_, msgData, err := a.conn.Read(a.ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Println("[AlchemyClient] Read loop stopped due to context cancellation")
			} else {
				log.Printf("[AlchemyClient] Read error: %v", err)
			}
			select {
			case a.errorChan <- err:
			default:
			}
			return
		}

		var msg SubscriptionResponse
		if err := json.Unmarshal(msgData, &msg); err != nil {
			log.Printf("JSON unmarshal error: %v", err)
			continue
		}
		if msg.Method != "eth_subscription" {
			continue
		}

		a.msgChan <- msg.Params.Result
	}
}
