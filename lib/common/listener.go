package common

import (
	"fmt"
	"log"
	"strings"

	"github.com/mi-yu/crypto-arbitrage/lib/config"
)

// Listener objects asynchronously receive quotes all exchanges
// and aggregates highest bids/lowest asks for given trading pair
type Listener struct {
	// Maps trading pair to quote channel
	pair           []string
	pairKey        string
	channelMap     map[string]chan *ExchangeQuote
	ps             *Pubsub
	bestAsk        *listenerQuote
	bestBid        *listenerQuote
	totalArbitrage float64
}

type listenerQuote struct {
	Price    float64
	Exchange string
}

func (lq *listenerQuote) String() string {
	return fmt.Sprintf("ListenerQuote{%s, %.6f}", lq.Exchange, lq.Price)
}

// NewListener creates a new Listener object, subscribing to updates from
// provided Pubsub
func NewListener(tradingPair []string, ps *Pubsub) *Listener {
	channelMap := make(map[string]chan *ExchangeQuote)
	pairKey := strings.Join(tradingPair, "-")
	for _, ex := range config.Exchanges {
		ch := make(chan *ExchangeQuote, 16)
		channelMap[ex] = ch
		ps.Subscribe(ex, pairKey, ch)
		fmt.Printf("Listener subscribed to exchange %s for pair %s%s\n", ex, tradingPair[0], tradingPair[1])
	}
	listener := &Listener{tradingPair, pairKey, channelMap, ps, nil, nil, 0.0}
	return listener
}

// Listen prints all current quotes across exchanges for
// tracked trading pair as they are published
func (lis *Listener) Listen() {
	for {
		for _, ex := range config.Exchanges {
			select {
			case quote := <-lis.channelMap[ex]:
				if lis.bestAsk == nil || quote.Ask < lis.bestAsk.Price || quote.Exchange == lis.bestAsk.Exchange {
					lis.bestAsk = &listenerQuote{Price: quote.Ask, Exchange: quote.Exchange}
				}
				if lis.bestBid == nil || quote.Bid > lis.bestBid.Price || quote.Exchange == lis.bestBid.Exchange {
					lis.bestBid = &listenerQuote{Price: quote.Bid, Exchange: quote.Exchange}
				}
				if lis.bestBid.Price > lis.bestAsk.Price {
					log.Printf("Arbitrage opportunity found: buy %s sell %s\n", lis.bestAsk.String(), lis.bestBid.String())
					lis.totalArbitrage += lis.bestBid.Price - lis.bestAsk.Price
					log.Printf("Total arbitrage: %.6f\n", lis.totalArbitrage)
				}
				// log.Printf("[%s-%s] :::: %s\n", ex, lis.pairKey, quote.String())
			default:
				// fmt.Println("No quote available from " + ex)
			}
		}
	}
}
