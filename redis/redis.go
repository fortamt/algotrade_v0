package redisstore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go-trader-bot/utils"
)

var ctx = context.Background()

type CandleStore struct {
	client *redis.Client
}

func NewCandleStore(addr string) *CandleStore {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr, // "localhost:6379"
	})
	return &CandleStore{client: rdb}
}

func redisKey(symbol, interval string) string {
	return fmt.Sprintf("candles:%s:%s", symbol, interval)
}

// Сохраняем новую свечу в Redis (список из 500)
func (cs *CandleStore) SaveCandle(symbol, interval string, k utils.Kline) error {
	key := redisKey(symbol, interval)
	data, err := json.Marshal(k)
	if err != nil {
		return err
	}
	pipe := cs.client.Pipeline()
	pipe.LPush(ctx, key, data)
	pipe.LTrim(ctx, key, 0, 499)
	_, err = pipe.Exec(ctx)
	return err
}

// Получаем последние N свечей
func (cs *CandleStore) GetLastN(symbol, interval string, n int) ([]utils.Kline, error) {
	key := redisKey(symbol, interval)
	values, err := cs.client.LRange(ctx, key, 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}

	result := make([]utils.Kline, 0, len(values))
	for _, v := range values {
		var k utils.Kline
		if err := json.Unmarshal([]byte(v), &k); err == nil {
			result = append(result, k)
		}
	}
	// В Redis самые новые свечи в начале — разворачиваем
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result, nil
}
