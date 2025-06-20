package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// Это будет WebSocket-адрес Binance Futures Testnet
// Мы подписываемся на 1-минутные свечи по BTCUSDT
const wsURL = "wss://fstream.binance.com/ws/btcusdt@kline_1m"

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

	// Читаем сообщения от Binance в отдельной горутине
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Ошибка чтения сообщения:", err)
				return
			}

			var data struct {
				Kline struct {
					Close   string `json:"c"`
					IsFinal bool   `json:"x"`
				} `json:"k"`
			}

			if err := json.Unmarshal(message, &data); err != nil {
				log.Println("Ошибка парсинга JSON:", err)
				continue
			}

			if data.Kline.IsFinal {
				fmt.Printf("Закрытие свечи: %s | Закрыта: %v\n", data.Kline.Close, data.Kline.IsFinal)
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
