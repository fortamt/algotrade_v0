package utils

import (
	"encoding/json"
	"strconv"
)

type Kline struct {
	OpenTime  int64   `json:"t"`
	Open      float64 `json:"o"`
	High      float64 `json:"h"`
	Low       float64 `json:"l"`
	Close     float64 `json:"c"`
	Volume    float64 `json:"v"`
	CloseTime int64   `json:"T"`
	IsFinal   bool    `json:"x"` // ← нужно обязательно
}

type RestKline struct {
	OpenTime  int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	CloseTime int64
}

type WebSocketKlineRaw struct {
	OpenTime  int64           `json:"t"`
	Open      json.RawMessage `json:"o"`
	High      json.RawMessage `json:"h"`
	Low       json.RawMessage `json:"l"`
	Close     json.RawMessage `json:"c"`
	Volume    json.RawMessage `json:"v"`
	CloseTime int64           `json:"T"`
	IsFinal   bool            `json:"x"`
}

func (r RestKline) ToKline() Kline {
	return Kline{
		OpenTime:  r.OpenTime,
		Open:      r.Open,
		High:      r.High,
		Low:       r.Low,
		Close:     r.Close,
		Volume:    r.Volume,
		CloseTime: r.CloseTime,
	}
}

func (r WebSocketKlineRaw) ToKline() (Kline, error) {
	return Kline{
		OpenTime:  r.OpenTime,
		Open:      ParseRawFloat(r.Open),
		High:      ParseRawFloat(r.High),
		Low:       ParseRawFloat(r.Low),
		Close:     ParseRawFloat(r.Close),
		Volume:    ParseRawFloat(r.Volume),
		CloseTime: r.CloseTime,
	}, nil
}

func ParseRawFloat(raw json.RawMessage) float64 {
	var f float64
	var s string

	// пробуем сначала как float
	if err := json.Unmarshal(raw, &f); err == nil {
		return f
	}

	// потом как строку
	if err := json.Unmarshal(raw, &s); err == nil {
		f, _ = strconv.ParseFloat(s, 64)
		return f
	}

	return 0
}
