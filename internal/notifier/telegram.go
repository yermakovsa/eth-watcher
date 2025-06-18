package notifier

import (
	"context"
	"fmt"

	"github.com/mymmrac/telego"
)

type Notifier interface {
	NotifyThresholdExceeded(ctx context.Context, txID, walletFrom string, walletTo string, total float64) error
}

type Bot interface {
	SendMessage(ctx context.Context, params *telego.SendMessageParams) (*telego.Message, error)
}

type TelegramNotifier struct {
	bot    Bot
	chatID int64
}

func NewTelegramNotifier(bot Bot, chatID int64) *TelegramNotifier {
	return &TelegramNotifier{
		bot:    bot,
		chatID: chatID,
	}
}

func (t *TelegramNotifier) NotifyThresholdExceeded(ctx context.Context, txID, walletFrom string, walletTo string, total float64) error {
	var msg string
	switch {
	case walletFrom != "":
		msg = fmt.Sprintf(
			"🔔 High Volume Detected\n\nSender: %s\nAmount: %.4f ETH\nTxID: %s",
			walletFrom, total, txID,
		)
	case walletTo != "":
		msg = fmt.Sprintf(
			"🔔 High Volume Detected\n\nReceiver: %s\nAmount: %.4f ETH\nTxID: %s",
			walletTo, total, txID,
		)
	default:
		return nil
	}

	params := &telego.SendMessageParams{}
	_, err := t.bot.SendMessage(ctx,
		params.
			WithChatID(telego.ChatID{ID: t.chatID}).
			WithText(msg),
	)

	return err
}
