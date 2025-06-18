package watcher

import (
	"context"
	"log"
	"strings"

	"github.com/yermakovsa/eth-watcher/internal/aggregator"
	"github.com/yermakovsa/eth-watcher/pkg/alchemy"
)

type AlchemyClient interface {
	SubscribeMinedTransactions(opts alchemy.MinedTxOptions) (<-chan alchemy.MinedTxEvent, error)
	Close() error
}

type Aggregator interface {
	Process(tx alchemy.MinedTxEvent, direction aggregator.Direction)
}

type Watcher struct {
	client      AlchemyClient
	aggregator  Aggregator
	walletsFrom map[string]struct{}
	walletsTo   map[string]struct{}
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewWatcher initializes a new transaction watcher
func NewWatcher(ctx context.Context, client AlchemyClient, from []string, to []string, aggregator Aggregator) *Watcher {
	w := &Watcher{
		client:      client,
		aggregator:  aggregator,
		walletsFrom: toSet(from),
		walletsTo:   toSet(to),
	}
	w.ctx, w.cancel = context.WithCancel(ctx)
	return w
}

// Start begins watching for mined transactions
func (w *Watcher) Start() error {
	var filters []alchemy.AddressFilter
	for wallet := range w.walletsFrom {
		filters = append(filters, alchemy.AddressFilter{From: wallet})
	}
	for wallet := range w.walletsTo {
		filters = append(filters, alchemy.AddressFilter{To: wallet})
	}

	events, err := w.client.SubscribeMinedTransactions(alchemy.MinedTxOptions{
		Addresses:      filters,
		IncludeRemoved: false,
		HashesOnly:     false,
	})
	if err != nil {
		return err
	}

	log.Println("[Watcher] Started transaction watcher")

	go w.watch(events)

	return nil
}

// Stop halts the watcher
func (w *Watcher) Stop() {
	log.Println("[Watcher] Stopping watcher")
	w.cancel()
	_ = w.client.Close()
}

func (w *Watcher) watch(events <-chan alchemy.MinedTxEvent) {
	for {
		select {
		case <-w.ctx.Done():
			log.Println("[Watcher] Shutdown signal received")
			return
		case event := <-events:
			from := strings.ToLower(event.Transaction.From)
			to := strings.ToLower(event.Transaction.To)

			if _, ok := w.walletsFrom[from]; ok {
				go w.aggregator.Process(event, aggregator.From)
			}
			if _, ok := w.walletsTo[to]; ok {
				go w.aggregator.Process(event, aggregator.To)
			}
		}
	}
}

// toSet converts a slice of wallet addresses to a normalized set (map for fast lookup)
func toSet(addresses []string) map[string]struct{} {
	set := make(map[string]struct{}, len(addresses))
	for _, addr := range addresses {
		set[strings.ToLower(addr)] = struct{}{}
	}
	return set
}
