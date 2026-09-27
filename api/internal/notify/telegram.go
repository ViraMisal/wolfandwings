package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Telegram — клиент Bot API. Отправка неблокирующая: NotifyAsync уходит в горутину.
type Telegram struct {
	BotToken string
	ChatID   string
	Log      *slog.Logger
}

// NotifyAsync отправляет сообщение в фоновой горутине; ошибки логируются.
func (t *Telegram) NotifyAsync(text string) {
	if t == nil || t.BotToken == "" || t.ChatID == "" || t.Log == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := t.send(ctx, text); err != nil {
			t.Log.Warn("tg notify", "err", err)
		}
	}()
}

func (t *Telegram) send(ctx context.Context, text string) error {
	body, err := json.Marshal(map[string]string{"chat_id": t.ChatID, "text": text})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.telegram.org/bot"+t.BotToken+"/sendMessage", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("tg api: %d", resp.StatusCode)
	}
	return nil
}
