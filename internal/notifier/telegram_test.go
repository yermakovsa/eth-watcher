package notifier

import (
	"context"
	"errors"
	"testing"

	"github.com/mymmrac/telego"
	"github.com/stretchr/testify/assert"
)

type mockBot struct {
	sendCalled bool
	shouldFail bool
}

func (m *mockBot) SendMessage(ctx context.Context, params *telego.SendMessageParams) (*telego.Message, error) {
	m.sendCalled = true
	if m.shouldFail {
		return nil, errors.New("send failed")
	}
	return &telego.Message{}, nil
}

func TestNotifyThresholdExceeded_Success(t *testing.T) {
	mock := &mockBot{}
	notifier := &TelegramNotifier{
		bot:    mock,
		chatID: 123456,
	}

	err := notifier.NotifyThresholdExceeded(context.Background(), "0xtxhash", "0xwallet", "", 123.45)

	assert.NoError(t, err)
	assert.True(t, mock.sendCalled)
}

func TestNotifyThresholdExceeded_Error(t *testing.T) {
	mock := &mockBot{shouldFail: true}
	notifier := &TelegramNotifier{
		bot:    mock,
		chatID: 123456,
	}

	err := notifier.NotifyThresholdExceeded(context.Background(), "0xtxhash", "0xwallet", "", 123.45)

	assert.Error(t, err)
	assert.True(t, mock.sendCalled)
}
