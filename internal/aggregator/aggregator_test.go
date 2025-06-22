package aggregator

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yermakovsa/alchemyws"
)

type MockNotifier struct {
	mu     sync.Mutex
	called bool
	args   struct {
		hash       string
		walletFrom string
		walletTo   string
		amount     float64
	}
}

func (m *MockNotifier) NotifyThresholdExceeded(ctx context.Context, txID, walletFrom string, walletTo string, total float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called = true
	m.args.hash = txID
	m.args.walletFrom = walletFrom
	m.args.walletTo = walletTo
	m.args.amount = total
	return nil
}

func TestAggregator_TriggerAlertWhenThresholdExceededFromWallet(t *testing.T) {
	notifier := &MockNotifier{}
	ctx := context.Background()

	agg := NewAggregator(ctx, notifier, 1.0, 10*time.Second, 5*time.Second)

	tx := alchemyws.MinedTxEvent{
		Transaction: alchemyws.Transaction{
			Hash:  "0x123",
			From:  "0xabc",
			Value: "0xde0b6b3a7640000", // 1 ETH
		},
	}

	go agg.Process(tx, From)

	time.Sleep(10 * time.Millisecond) // wait for goroutine

	notifier.mu.Lock()
	defer notifier.mu.Unlock()
	assert.True(t, notifier.called)
	assert.Equal(t, "0x123", notifier.args.hash)
	assert.Equal(t, "0xabc", notifier.args.walletFrom)
	assert.Equal(t, "", notifier.args.walletTo)
	assert.InDelta(t, 1.0, notifier.args.amount, 0.0001)
}

func TestAggregator_TriggerAlertWhenThresholdExceededToWallet(t *testing.T) {
	notifier := &MockNotifier{}
	ctx := context.Background()

	agg := NewAggregator(ctx, notifier, 1.0, 10*time.Second, 5*time.Second)

	tx := alchemyws.MinedTxEvent{
		Transaction: alchemyws.Transaction{
			Hash:  "0x123",
			To:    "0xabc",
			Value: "0xde0b6b3a7640000", // 1 ETH
		},
	}

	go agg.Process(tx, To)

	time.Sleep(10 * time.Millisecond) // wait for goroutine

	notifier.mu.Lock()
	defer notifier.mu.Unlock()
	assert.True(t, notifier.called)
	assert.Equal(t, "0x123", notifier.args.hash)
	assert.Equal(t, "0xabc", notifier.args.walletTo)
	assert.Equal(t, "", notifier.args.walletFrom)
	assert.InDelta(t, 1.0, notifier.args.amount, 0.0001)
}

func TestAggregator_DoesNotTriggerAlertBelowThreshold(t *testing.T) {
	notifier := &MockNotifier{}
	ctx := context.Background()

	agg := NewAggregator(ctx, notifier, 2.0, 10*time.Second, 5*time.Second)

	tx := alchemyws.MinedTxEvent{
		Transaction: alchemyws.Transaction{
			Hash:  "0x456",
			From:  "0xabc",
			Value: "0xde0b6b3a7640000", // 1 ETH
		},
	}

	go agg.Process(tx, From)
	go agg.Process(tx, To)

	time.Sleep(10 * time.Millisecond)

	notifier.mu.Lock()
	defer notifier.mu.Unlock()
	assert.False(t, notifier.called)
}

func TestAggregator_CooldownPreventsMultipleAlerts(t *testing.T) {
	notifier := &MockNotifier{}
	ctx := context.Background()

	agg := NewAggregator(ctx, notifier, 1.0, 10*time.Second, 1*time.Second)

	tx := alchemyws.MinedTxEvent{
		Transaction: alchemyws.Transaction{
			Hash:  "0x789",
			From:  "0xabc",
			Value: "0xde0b6b3a7640000", // 1 ETH
		},
	}

	// First call - should alert
	go agg.Process(tx, From)
	time.Sleep(10 * time.Millisecond)

	// Second call within cooldown - should NOT alert again
	notifier.mu.Lock()
	notifier.called = false // reset flag
	notifier.mu.Unlock()

	go agg.Process(tx, From)
	time.Sleep(10 * time.Millisecond)

	notifier.mu.Lock()
	defer notifier.mu.Unlock()
	assert.False(t, notifier.called)
}

func TestParseValue_CorrectConversion(t *testing.T) {
	eth := ParseValue("0xde0b6b3a7640000") // 1 ETH
	assert.InDelta(t, 1.0, eth, 0.00001)

	halfEth := ParseValue("0x1bc16d674ec80000") // 2 ETH
	assert.InDelta(t, 2, halfEth, 0.00001)

	invalid := ParseValue("nothex")
	assert.Equal(t, 0.0, invalid)
}
