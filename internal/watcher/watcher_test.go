package watcher_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yermakovsa/eth-watcher/internal/aggregator"
	"github.com/yermakovsa/eth-watcher/internal/watcher"
	"github.com/yermakovsa/eth-watcher/pkg/alchemy"
)

type MockAlchemyClient struct {
	SubscribeMinedTransactionsFunc func(opts alchemy.MinedTxOptions) (<-chan alchemy.MinedTxEvent, error)
	CloseFunc                      func() error
}

func (m *MockAlchemyClient) SubscribeMinedTransactions(opts alchemy.MinedTxOptions) (<-chan alchemy.MinedTxEvent, error) {
	return m.SubscribeMinedTransactionsFunc(opts)
}
func (m *MockAlchemyClient) Close() error {
	return m.CloseFunc()
}

type MockAggregator struct {
	ProcessFunc func(event alchemy.MinedTxEvent, direction aggregator.Direction)
}

func (m *MockAggregator) Process(event alchemy.MinedTxEvent, direction aggregator.Direction) {
	if m.ProcessFunc != nil {
		m.ProcessFunc(event, direction)
	}
}

func TestWatcher_Start_ProcessesFromWalletEventAndStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eventReceived := make(chan alchemy.MinedTxEvent, 1)
	eventDirectionReceived := make(chan aggregator.Direction, 1)

	mockAggregator := &MockAggregator{
		ProcessFunc: func(e alchemy.MinedTxEvent, direction aggregator.Direction) {
			eventReceived <- e
			eventDirectionReceived <- direction
		},
	}

	events := make(chan alchemy.MinedTxEvent, 1)
	events <- alchemy.MinedTxEvent{Transaction: alchemy.MinedTransaction{From: "0xabc"}}

	mockClient := &MockAlchemyClient{
		SubscribeMinedTransactionsFunc: func(opts alchemy.MinedTxOptions) (<-chan alchemy.MinedTxEvent, error) {
			return events, nil
		},
		CloseFunc: func() error { return nil },
	}

	w := watcher.NewWatcher(ctx, mockClient, []string{"0xabc"}, []string{}, mockAggregator)

	errW := w.Start()
	assert.NoError(t, errW)

	select {
	case e := <-eventReceived:
		assert.Equal(t, "0xabc", e.Transaction.From)
	case <-time.After(1 * time.Second):
		t.Fatal("expected transaction event not received")
	}

	select {
	case e := <-eventDirectionReceived:
		assert.Equal(t, aggregator.From, e)
	case <-time.After(1 * time.Second):
		t.Fatal("expected transaction event not received")
	}

	cancel() // Clean shutdown
}

func TestWatcher_Start_ProcessesToWalletEventAndStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eventReceived := make(chan alchemy.MinedTxEvent, 1)
	eventDirectionReceived := make(chan aggregator.Direction, 1)

	mockAggregator := &MockAggregator{
		ProcessFunc: func(e alchemy.MinedTxEvent, direction aggregator.Direction) {
			eventReceived <- e
			eventDirectionReceived <- direction
		},
	}

	events := make(chan alchemy.MinedTxEvent, 1)
	events <- alchemy.MinedTxEvent{Transaction: alchemy.MinedTransaction{To: "0xabc"}}

	mockClient := &MockAlchemyClient{
		SubscribeMinedTransactionsFunc: func(opts alchemy.MinedTxOptions) (<-chan alchemy.MinedTxEvent, error) {
			return events, nil
		},
		CloseFunc: func() error { return nil },
	}

	w := watcher.NewWatcher(ctx, mockClient, []string{}, []string{"0xabc"}, mockAggregator)

	errW := w.Start()
	assert.NoError(t, errW)

	select {
	case e := <-eventReceived:
		assert.Equal(t, "0xabc", e.Transaction.To)
	case <-time.After(1 * time.Second):
		t.Fatal("expected transaction event not received")
	}

	select {
	case e := <-eventDirectionReceived:
		assert.Equal(t, aggregator.To, e)
	case <-time.After(1 * time.Second):
		t.Fatal("expected transaction event not received")
	}

	cancel() // Clean shutdown
}

func TestWatcher_Start_SubscriptionFails(t *testing.T) {
	mockClient := &MockAlchemyClient{
		SubscribeMinedTransactionsFunc: func(opts alchemy.MinedTxOptions) (<-chan alchemy.MinedTxEvent, error) {
			return nil, errors.New("subscription failed")
		},
		CloseFunc: func() error { return nil },
	}

	mockAggregator := &MockAggregator{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := watcher.NewWatcher(ctx, mockClient, []string{"0xabc"}, []string{}, mockAggregator)

	errW := w.Start()
	assert.Error(t, errW)
	assert.Contains(t, errW.Error(), "subscription failed")
}

func TestWatcher_Stop_ClosesClient(t *testing.T) {
	closed := false

	mockClient := &MockAlchemyClient{
		SubscribeMinedTransactionsFunc: func(opts alchemy.MinedTxOptions) (<-chan alchemy.MinedTxEvent, error) {
			return make(chan alchemy.MinedTxEvent), nil
		},
		CloseFunc: func() error {
			closed = true
			return nil
		},
	}

	mockAggregator := &MockAggregator{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := watcher.NewWatcher(ctx, mockClient, []string{"0xabc"}, []string{}, mockAggregator)

	go func() {
		_ = w.Start()
	}()

	time.Sleep(100 * time.Millisecond) // Let it start

	w.Stop()

	assert.True(t, closed, "expected client to be closed on Stop()")
}
