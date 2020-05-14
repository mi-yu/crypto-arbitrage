package exchanges

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/mi-yu/crypto-arbitrage/lib/common"
	"github.com/mi-yu/crypto-arbitrage/lib/config"

	"github.com/gorilla/websocket"
)

const coinbaseFee float64 = 0.0149
const coinbaseEndpoint string = "wss://ws-feed.pro.coinbase.com"

// CoinbaseQuote represents a single websocket ticker
// message from Coinbase API
type CoinbaseQuote struct {
	Type      string  `json:"type"`
	TradeID   int     `json:"trade_id,string"`
	Sequence  int     `json:"sequence,string"`
	Time      string  `json:"time"`
	ProductID string  `json:"product_id"`
	Price     float64 `json:"price,string"`
	Side      string  `json:"buy"`
	LastSize  float64 `json:"last_size,string"`
	BestBid   float64 `json:"best_bid,string"`
	BestAsk   float64 `json:"best_ask,string"`
}

// CoinbaseExchange is a wrapper for interacting with
// the Coinbase exchange
type CoinbaseExchange struct {
	name   string
	pubsub *common.Pubsub
}

// CoinbaseWSSubscriptionMsg is the structure of message
// sent to subscribe to a Coinbase websocket
type CoinbaseWSSubscriptionMsg struct {
	Type       string   `json:"type"`
	ProductIDs []string `json:"product_ids"`
	Channels   []string `json:"channels"`
}

// NewCoinbaseExchange creates a new CoinbaseExchange instance
func NewCoinbaseExchange(ps *common.Pubsub) *CoinbaseExchange {
	ex := &CoinbaseExchange{"coinbase", ps}
	return ex
}

// Start begins listening to quotes from Coinbase
func (ex *CoinbaseExchange) Start() {
	var subscriptions []string
	for _, pair := range config.TradingPairs {
		subscriptions = append(subscriptions, fmt.Sprintf("%s-%s", pair[0], pair[1]))
	}
	c, _, err := websocket.DefaultDialer.Dial(coinbaseEndpoint, nil)

	if err != nil {
		log.Fatal("Failed to dial Coinbase websocket: ", err)
	}
	subMsg, _ := json.Marshal(&CoinbaseWSSubscriptionMsg{
		Type:       "subscribe",
		ProductIDs: subscriptions,
		Channels:   []string{"ticker"},
	})
	if err != nil {
		log.Fatal("Failed to subscribe to Coinbase: ", err)
	}
	c.WriteMessage(websocket.TextMessage, subMsg)
	defer c.Close()

	done := make(chan struct{})

	// Handling incoming messages from websocket
	go func() {
		defer close(done)
		for {
			var quote CoinbaseQuote
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			json.Unmarshal(msg, &quote)

			quote.BestAsk *= (1.0 + coinbaseFee)
			quote.BestBid *= (1.0 - coinbaseFee)

			ex.pubsub.Publish(
				ex.name,
				quote.ProductID,
				&common.ExchangeQuote{Bid: quote.BestBid, Ask: quote.BestAsk, Exchange: ex.name},
			)
		}
	}()
	for {
		select {
		case <-done:
			return

		}
	}
}
