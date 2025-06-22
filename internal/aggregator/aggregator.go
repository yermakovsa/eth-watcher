package aggregator

import (
	"context"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/yermakovsa/alchemyws"
	"github.com/yermakovsa/eth-watcher/internal/notifier"
)

type Direction string

const (
	From Direction = "from"
	To   Direction = "to"
)

type TxRecord struct {
	Amount    float64
	Timestamp time.Time
}

// Aggregator monitors wallet activity and triggers alerts when volume exceeds threshold.
type Aggregator struct {
	mu        sync.Mutex
	data      map[Direction]map[string][]TxRecord
	alerted   map[Direction]map[string]time.Time
	threshold float64
	window    time.Duration
	cooldown  time.Duration
	notifier  notifier.Notifier
	ctx       context.Context
}

// NewAggregator initializes an Aggregator.
func NewAggregator(ctx context.Context, notifier notifier.Notifier, threshold float64, window time.Duration, cooldown time.Duration) *Aggregator {
	return &Aggregator{
		data: map[Direction]map[string][]TxRecord{
			From: make(map[string][]TxRecord),
			To:   make(map[string][]TxRecord),
		},
		alerted: map[Direction]map[string]time.Time{
			From: make(map[string]time.Time),
			To:   make(map[string]time.Time),
		},
		threshold: threshold,
		window:    window,
		cooldown:  cooldown,
		notifier:  notifier,
		ctx:       ctx,
	}
}

// Process adds a transaction to the aggregation buffer and triggers alert if needed.
func (a *Aggregator) Process(tx alchemyws.MinedTxEvent, direction Direction) {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	amount := ParseValue(tx.Transaction.Value)
	var wallet string
	switch direction {
	case From:
		wallet = strings.ToLower(tx.Transaction.From)
	case To:
		wallet = strings.ToLower(tx.Transaction.To)
	default:
		return
	}

	// Append transaction
	a.data[direction][wallet] = append(a.data[direction][wallet], TxRecord{
		Amount:    amount,
		Timestamp: now,
	})

	// Filter transactions in the window
	var (
		recent []TxRecord
		total  float64
	)

	for _, r := range a.data[direction][wallet] {
		if now.Sub(r.Timestamp) <= a.window {
			recent = append(recent, r)
			total += r.Amount
		}
	}
	a.data[direction][wallet] = recent

	if total < a.threshold {
		return
	}

	lastAlert, alerted := a.alerted[direction][wallet]
	if alerted && now.Sub(lastAlert) <= a.cooldown {
		return
	}

	// Check alert condition
	a.alerted[direction][wallet] = now

	var walletFrom, walletTo string
	if direction == From {
		walletFrom = wallet
	} else {
		walletTo = wallet
	}

	go a.notifier.NotifyThresholdExceeded(a.ctx, tx.Transaction.Hash, walletFrom, walletTo, total)
}

func ParseValue(raw string) float64 {
	cleaned := strings.TrimPrefix(strings.ToLower(raw), "0x")
	bigVal, ok := new(big.Int).SetString(cleaned, 16)
	if !ok {
		log.Printf("Failed to parse value '%s' as hexadecimal", raw)
		return 0
	}

	wei := new(big.Float).SetInt(bigVal)
	eth := new(big.Float).Quo(wei, big.NewFloat(1e18))
	result, _ := eth.Float64()
	return result
}
