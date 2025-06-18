package alchemy

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
)

type MockConn struct {
	WriteFunc func(ctx context.Context, typ websocket.MessageType, p []byte) error
	ReadFunc  func(ctx context.Context) (websocket.MessageType, []byte, error)
	CloseFunc func(code websocket.StatusCode, reason string) error
}

func (m *MockConn) Write(ctx context.Context, typ websocket.MessageType, p []byte) error {
	return m.WriteFunc(ctx, typ, p)
}

func (m *MockConn) Read(ctx context.Context) (websocket.MessageType, []byte, error) {
	return m.ReadFunc(ctx)
}

func (m *MockConn) Close(code websocket.StatusCode, reason string) error {
	return m.CloseFunc(code, reason)
}

func TestSubscribeMinedTransactions_Success(t *testing.T) {
	called := false

	mock := &MockConn{
		WriteFunc: func(ctx context.Context, typ websocket.MessageType, p []byte) error {
			return nil
		},
		ReadFunc: func(ctx context.Context) (websocket.MessageType, []byte, error) {
			if called {
				// Block so the loop doesn't exit early, but also doesn't send more messages
				<-ctx.Done()
				return 0, nil, ctx.Err()
			}
			called = true

			msg := SubscriptionResponse{
				Method: "eth_subscription",
				Params: SubscriptionBody{
					Result: MinedTxEvent{
						Transaction: MinedTransaction{From: "0xabc"},
					},
				},
			}
			data, _ := json.Marshal(msg)
			return websocket.MessageText, data, nil
		},
		CloseFunc: func(code websocket.StatusCode, reason string) error {
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := NewAlchemyClientWithConn(ctx, mock)

	ch, err := client.SubscribeMinedTransactions(MinedTxOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	select {
	case tx := <-ch:
		assert.Equal(t, "0xabc", tx.Transaction.From)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected mined transaction not received")
	}
}

func TestSubscribeMinedTransactions_JSONError(t *testing.T) {
	called := false
	mock := &MockConn{
		WriteFunc: func(ctx context.Context, typ websocket.MessageType, p []byte) error {
			return nil
		},
		ReadFunc: func(ctx context.Context) (websocket.MessageType, []byte, error) {
			if called {
				// Block so the loop doesn't exit early, but also doesn't send more messages
				<-ctx.Done()
				return 0, nil, ctx.Err()
			}
			called = true
			return websocket.MessageText, []byte(`invalid-json`), nil
		},
		CloseFunc: func(code websocket.StatusCode, reason string) error {
			return nil
		},
	}

	client := NewAlchemyClientWithConn(context.Background(), mock)
	ch, err := client.SubscribeMinedTransactions(MinedTxOptions{})
	assert.NoError(t, err)

	select {
	case <-ch:
		t.Fatal("expected no valid message")
	case <-time.After(500 * time.Millisecond):
		// passed
	}
}

func TestSubscribeMinedTransactions_ReadError(t *testing.T) {
	mock := &MockConn{
		WriteFunc: func(ctx context.Context, typ websocket.MessageType, p []byte) error {
			return nil
		},
		ReadFunc: func(ctx context.Context) (websocket.MessageType, []byte, error) {
			return 0, nil, errors.New("read failed")
		},
		CloseFunc: func(code websocket.StatusCode, reason string) error {
			return nil
		},
	}

	client := NewAlchemyClientWithConn(context.Background(), mock)
	_, err := client.SubscribeMinedTransactions(MinedTxOptions{})
	assert.NoError(t, err)

	select {
	case err := <-client.errorChan:
		assert.EqualError(t, err, "read failed")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected error not received")
	}
}

func TestAlchemyClient_Close(t *testing.T) {
	mock := &MockConn{
		CloseFunc: func(code websocket.StatusCode, reason string) error {
			assert.Equal(t, websocket.StatusNormalClosure, code)
			assert.Equal(t, "client closed", reason)
			return nil
		},
	}

	client := NewAlchemyClientWithConn(context.Background(), mock)
	err := client.Close()
	assert.NoError(t, err)
}
