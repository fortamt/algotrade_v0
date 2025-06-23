package strategy

import "go-trader-bot/utils"

type SMA struct{}

func (s SMA) Name() string {
	return "SMA"
}

func (s SMA) Run(candles []utils.Kline) string {
	if len(candles) < 11 {
		return "none"
	}

	var c5, c10 []float64
	for i := len(candles) - 5; i < len(candles); i++ {
		c5 = append(c5, candles[i].Close)
	}
	for i := len(candles) - 10; i < len(candles); i++ {
		c10 = append(c10, candles[i].Close)
	}

	sma5 := average(c5)
	sma10 := average(c10)

	if sma5 > sma10 {
		return "buy"
	} else if sma5 < sma10 {
		return "sell"
	}
	return "none"
}

func average(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
