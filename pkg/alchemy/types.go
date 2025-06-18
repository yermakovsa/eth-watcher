package alchemy

// MinedTxOptions defines the parameters for the subscription request
type MinedTxOptions struct {
	Addresses      []AddressFilter `json:"addresses,omitempty"`
	IncludeRemoved bool            `json:"includeRemoved,omitempty"`
	HashesOnly     bool            `json:"hashesOnly,omitempty"`
}

type AddressFilter struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

// SubscriptionRequest is the structure of the eth_subscribe call
type SubscriptionRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

// SubscriptionResponse is the main message response wrapper
type SubscriptionResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  SubscriptionBody `json:"params"`
}

// SubscriptionBody wraps the result and subscription ID
type SubscriptionBody struct {
	Result       MinedTxEvent `json:"result"`
	Subscription string       `json:"subscription"`
}

// MinedTxEvent holds the actual transaction data and if it's removed
type MinedTxEvent struct {
	Removed     bool             `json:"removed"`
	Transaction MinedTransaction `json:"transaction"`
}

// MinedTransaction is a simplified version of Ethereum transaction
type MinedTransaction struct {
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"`
	Hash             string `json:"hash"`
	From             string `json:"from"`
	To               string `json:"to"`
	Value            string `json:"value"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Nonce            string `json:"nonce"`
	TransactionIndex string `json:"transactionIndex"`
}
