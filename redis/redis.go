package redisstore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go-trader-bot/utils"
)

type CandleStore struct {
	client *redis.Client
}

var ctx = context.Background()

func NewClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func NewCandleStoreFromClient(rdb *redis.Client) *CandleStore {
	return &CandleStore{client: rdb}
}

func (cs *CandleStore) SaveCandle(symbol, interval string, kline utils.Kline) error {
	key := fmt.Sprintf("candles:%s:%s", symbol, interval)

	data, err := json.Marshal(kline)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	pipe := cs.client.TxPipeline()
	pipe.LPush(ctx, key, data)
	pipe.LTrim(ctx, key, 0, 499)
	pipe.Publish(ctx, "candles:new", fmt.Sprintf("%s:%s", symbol, interval))
	_, err = pipe.Exec(ctx)
	return err
}

func (cs *CandleStore) GetLastN(symbol, interval string, n int) ([]utils.Kline, error) {
	key := fmt.Sprintf("candles:%s:%s", symbol, interval)

	values, err := cs.client.LRange(ctx, key, 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}

	candles := make([]utils.Kline, 0, len(values))
	for i := len(values) - 1; i >= 0; i-- {
		var k utils.Kline
		if err := json.Unmarshal([]byte(values[i]), &k); err != nil {
			continue
		}
		candles = append(candles, k)
	}

	return candles, nil
}
