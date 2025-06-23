package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"go-trader-bot/redis"
	"go-trader-bot/strategy"
	"go-trader-bot/utils"

	"github.com/gorilla/websocket"
)

// Это будет WebSocket-адрес Binance Futures Testnet
// Мы подписываемся на 1-минутные свечи по BTCUSDT

const (
	wsURL    = "wss://fstream.binance.com/ws/btcusdt@kline_1m"
	symbol   = "BTCUSDT"
	interval = "1m"
)

var limit = 500

func main() {
	strategy.Register(&strategy.Breakout{})
	go strategy.StartStrategyWorker("localhost:6379")
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

	rdb := redisstore.NewClient("localhost:6379")
	store := redisstore.NewCandleStoreFromClient(rdb)

	restCandles, err := utils.GetRecentCandles(symbol, interval, limit)
	if err != nil {
		log.Fatalf("REST load failed: %v", err)
	}

	for _, c := range restCandles {
		if err := store.SaveCandle(symbol, interval, c); err != nil {
			log.Println("Redis write error:", err)
		}
	}

	log.Printf("Инициализирован Redis-буфер: %d свечей", len(restCandles))

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
				K utils.WebSocketKlineRaw `json:"k"`
			}

			if err := json.Unmarshal(message, &payload); err != nil {
				log.Println("Ошибка парсинга JSON:", err)
				continue
			}

			if payload.K.IsFinal {
				k, _ := payload.K.ToKline()
				if err := store.SaveCandle(symbol, interval, k); err != nil {
					log.Println("Redis write error:", err)
				} else {
					fmt.Printf("Redis updated: close=%.2f\n", k.Close)
				}
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
