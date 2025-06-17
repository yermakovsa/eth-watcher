# Ethereum Transaction Watcher

A minimal and efficient Go-based service for monitoring Ethereum wallet activity in real time.  
It aggregates transactions using the Alchemy WebSocket API and triggers Telegram alerts when a configurable ETH transfer threshold is exceeded.

This project is built for observability, extensibility, and low operational overhead.

## Features

- 🔄 Real-time Ethereum transaction monitoring via Alchemy WebSocket API
- 📈 Aggregation of transaction volumes over configurable time windows
- 🚨 Telegram notifications for high-volume wallet activity
- 🔍 Separate tracking for `wallets from` and `wallets to`
- ⚙️ Customizable thresholds, cooldown periods, and monitored addresses
- 🧪 Built with modularity in mind - easily extendable for other notifiers or chains

## Getting Started

### Prerequisites

- Go 1.21 or later
- An [Alchemy](https://www.alchemy.com/) API key
- A Telegram bot token and chat ID (see [BotFather](https://telegram.me/BotFather))

### Installation

Clone the repository:

```bash
git clone https://github.com/yermakovsa/eth-watcher.git
cd eth-watcher
```

Install dependencies:

```bash
go mod tidy
```

### Configuration

Create a `.env` file in the project root with the following variables:

```env
# Alchemy API Key
ALCHEMY_API_KEY=your-alchemy-api-key

# Telegram Bot configuration
TELEGRAM_BOT_API_KEY=your-telegram-bot-token
TELEGRAM_CHAT_ID=your-chat-id

# Monitored wallet addresses (comma-separated)
MONITORED_WALLETS_FROM=0xabc...,0xdef...
MONITORED_WALLETS_TO=0x123...,0x456...

# Aggregation settings
AGGREGATION_WINDOW_IN_SECONDS=300                 # Time window for aggregation (in seconds)
AGGREGATION_NOTIFICATION_COOLDOWN_IN_SECONDS=60   # Min interval between repeated alerts
THRESHOLD_ETH=10.0                                # Volume threshold (in ETH) to trigger alert
```

## Running the Application

### Prerequisites

- Go 1.21+
- `.env` file properly configured (see Configuration)

### Run

```bash
go run ./cmd/app/main.go
```

The application will:

* Subscribe to new mined transactions using the Alchemy WebSocket API

* Monitor specified from and / or to wallet addresses

* Aggregate transaction volume within the configured time window

* Send a Telegram alert if the volume exceeds the configured threshold

## Running with Docker

You can run the application inside a Docker container for easier deployment.

### Build the Docker Image

```bash
docker build -t eth-watcher .
```

### Run the Container

Use environment variables to configure the container:

```bash
docker run --rm \
  -e ALCHEMY_API_KEY=your-alchemy-key \
  -e TELEGRAM_BOT_API_KEY=your-telegram-bot-key \
  -e TELEGRAM_CHAT_ID=your-chat-id \
  -e MONITORED_WALLETS_FROM=0xWallet1,0xWallet2 \
  -e MONITORED_WALLETS_TO=0xWallet3 \
  -e AGGREGATION_WINDOW_IN_SECONDS=300 \
  -e AGGREGATION_NOTIFICATION_COOLDOWN_IN_SECONDS=30 \
  -e THRESHOLD_ETH=0.0 \
  eth-watcher
```

### Required Parameters

The following environment variables must be set:

* `ALCHEMY_API_KEY`
* `TELEGRAM_BOT_API_KEY`
* `TELEGRAM_CHAT_ID`
* At least one of:
   * `MONITORED_WALLETS_FROM`
   * `MONITORED_WALLETS_TO`

### Optional Parameters (defaults shown)
* `AGGREGATION_WINDOW_IN_SECONDS` — default: 300
* `AGGREGATION_NOTIFICATION_COOLDOWN_IN_SECONDS` — default: 30
* `THRESHOLD_ETH` — default: 0.0

## License

This project is licensed under the [MIT License](LICENSE).

You are free to use, modify, and distribute this software in both personal and commercial projects, as long as you include the original license and copyright

