package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go-trader-bot/utils"
)

// Это будет WebSocket-адрес Binance Futures Testnet
// Мы подписываемся на 1-минутные свечи по BTCUSDT

const (
	wsURL    = "wss://fstream.binance.com/ws/btcusdt@kline_1m"
	symbol   = "BTCUSDT"
	interval = "1m"
)

var limit = 500

var (
	candleBuffer []utils.Kline
	mu           sync.Mutex
)

func main() {
	// Это нужно, чтобы корректно завершать программу по Ctrl+C
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Подключаемся к WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("Ошибка подключения к WebSocket:", err)
	}
	defer conn.Close()

	// Канал для завершения чтения
	done := make(chan struct{})

	initialCandles, err := utils.GetRecentCandles(symbol, interval, limit)
	if err != nil {
		log.Fatalf("Failed to load initial candles: %v", err)
	}

	candleBuffer = initialCandles
	log.Printf("Loaded %d candles", len(candleBuffer))

	// Читаем сообщения от Binance в отдельной горутине
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Ошибка чтения сообщения:", err)
				return
			}

			var payload struct {
				K utils.Kline `json:"k"`
			}

			if err := json.Unmarshal(message, &payload); err != nil {
				log.Println("Ошибка парсинга JSON:", err)
				continue
			}

			if payload.K.IsFinal {
				mu.Lock()
				if len(candleBuffer) >= limit {
					candleBuffer = candleBuffer[1:]
				}
				candleBuffer = append(candleBuffer, payload.K)
				mu.Unlock()

				fmt.Printf("Новая закрытая свеча: close=%v, total=%d\n", payload.K.Close, len(candleBuffer))
			}
		}
	}()

	// Основной цикл программы — ждёт завершения или Ctrl+C
	for {
		select {
		case <-done:
			// Если чтение закрылось — выходим
			return
		case <-interrupt:
			log.Println("Отключение...")

			// Отправляем серверу корректный запрос на закрытие
			err := conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "выход"))
			if err != nil {
				log.Println("Ошибка отправки close-сообщения:", err)
				return
			}

			// Даём немного времени серверу на ответ и выходим
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
