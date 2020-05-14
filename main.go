package main

import (
	"github.com/mi-yu/crypto-arbitrage/lib/config"

	"github.com/mi-yu/crypto-arbitrage/lib/exchanges"

	"github.com/mi-yu/crypto-arbitrage/lib/common"
)

func main() {
	ps := common.NewPubsub()
	exs := exchanges.InitExchanges(ps)

	var listeners []*common.Listener
	for _, pair := range config.TradingPairs {
		listeners = append(listeners, common.NewListener(pair, ps))
	}

	for _, ex := range exs {
		go ex.Start()
	}

	done := make(chan bool)

	for _, listener := range listeners {
		go listener.Listen()
	}

	for {
		select {
		case <-done:
			return
		}
	}
}
