package exchanges

import (
	"github.com/mi-yu/crypto-arbitrage/lib/config"

	"github.com/mi-yu/crypto-arbitrage/lib/common"
)

func exchangeToInterface(ex common.Exchange) common.Exchange {
	return ex
}

// InitExchanges creates instances for all exchanges
// and returns them as a list
func InitExchanges(ps *common.Pubsub) []common.Exchange {
	var exs []common.Exchange
	for _, ex := range config.Exchanges {
		switch ex {
		case "binance":
			exs = append(exs, exchangeToInterface(NewBinanceExchange(ps)))
		case "coinbase":
			exs = append(exs, exchangeToInterface(NewCoinbaseExchange(ps)))
		}
	}

	return exs
}
