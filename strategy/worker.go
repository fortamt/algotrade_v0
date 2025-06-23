package strategy

import (
	"context"
	"log"
	"strings"

	"go-trader-bot/redis"
)

var strategies []Strategy

func Register(s Strategy) {
	strategies = append(strategies, s)
}

func StartStrategyWorker(redisAddr string) {
	rdb := redisstore.NewClient(redisAddr)
	store := redisstore.NewCandleStoreFromClient(rdb)

	sub := rdb.Subscribe(context.Background(), "candles:new")
	ch := sub.Channel()

	for msg := range ch {
		parts := strings.Split(msg.Payload, ":")
		if len(parts) != 2 {
			continue
		}
		symbol, interval := parts[0], parts[1]

		candles, err := store.GetLastN(symbol, interval, 50)
		if err != nil {
			log.Println("Ошибка чтения свечей:", err)
			continue
		}

		for _, s := range strategies {
			signal := s.Run(candles)
			if signal != "none" {
				log.Printf("📊 %s [%s %s] → %s", s.Name(), symbol, interval, signal)
			} else {
				log.Printf("нет")
			}
		}
	}
}
