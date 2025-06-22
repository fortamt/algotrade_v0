package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func GetRecentCandles(symbol string, interval string, limit int) ([]Kline, error) {
	//для обработки случая одной незакрытой свечи
	fullLimit := limit + 1

	url := fmt.Sprintf("https://fapi.binance.com/fapi/v1/klines?symbol=%s&interval=%s&limit=%d", symbol, interval, fullLimit)

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer response.Body.Close()

	var raw [][]any

	if err := json.NewDecoder(response.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode error %w", err)
	}

	now := time.Now().UnixMilli()
	candles := make([]Kline, 0, limit)

	for i, rec := range raw {
		if len(rec) < 7 {
			continue // плохо сформированная свеча
		}

		closeTime := int64(rec[6].(float64))
		if i == len(raw)-1 && closeTime > now {
			// последняя свеча ещё не закрыта — пропускаем
			break
		}

		c := Kline{
			OpenTime:  int64(rec[0].(float64)),
			Open:      mustParseFloatFromString(rec[1]),
			High:      mustParseFloatFromString(rec[2]),
			Low:       mustParseFloatFromString(rec[3]),
			Close:     mustParseFloatFromString(rec[4]),
			Volume:    mustParseFloatFromString(rec[5]),
			CloseTime: closeTime,
		}
		candles = append(candles, c)
	}

	return candles, nil

}

func mustParseFloatFromString(v any) float64 {
	s, ok := v.(string)
	if !ok {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
