package utils

type Kline struct {
	Symbol    string
	Interval  string
	OpenTime  int64
	CloseTime int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	IsFinal   bool // только для WebSocket
}
