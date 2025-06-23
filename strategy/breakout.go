package strategy

import (
	"go-trader-bot/utils"
)

type Breakout struct {
	lastSignal string
}

func (b *Breakout) Name() string {
	return "Breakout-SMA"
}

func (b *Breakout) Run(candles []utils.Kline) string {
	if len(candles) < 20 {
		return "none"
	}

	// Диапазон по последним 20 свечам
	high := candles[0].High
	low := candles[0].Low
	sum := 0.0

	for _, c := range candles[len(candles)-20:] {
		if c.High > high {
			high = c.High
		}
		if c.Low < low {
			low = c.Low
		}
		sum += c.Close
	}

	sma := sum / 20
	latest := candles[len(candles)-1]

	var signal string
	if latest.Close > high && latest.Close > sma {
		signal = "buy"
	} else if latest.Close < low && latest.Close < sma {
		signal = "sell"
	} else {
		signal = "none"
	}

	if signal == b.lastSignal {
		return "none"
	}

	b.lastSignal = signal
	return signal
}
