package strategy

import "go-trader-bot/utils"

type Strategy interface {
	Name() string
	Run([]utils.Kline) string
}
