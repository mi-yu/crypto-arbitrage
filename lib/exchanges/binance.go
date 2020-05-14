package exchanges

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/mi-yu/crypto-arbitrage/lib/config"

	"github.com/mi-yu/crypto-arbitrage/lib/common"

	"github.com/gorilla/websocket"
)

const binanceFee float64 = 0.001
const binanceEndpoint string = "wss://stream.binance.com:9443"

// BinanceExchange provides interface for interacting with Binance
type BinanceExchange struct {
	name         string
	pubsub       *common.Pubsub
	pubsubKeyMap map[string]string
}

type BinanceWSPayload struct {
	Stream string       `json:"stream"`
	Data   BinanceQuote `json:"data"`
}

type BinanceQuote struct {
	UpdateID string  `json:"u"`
	Symbol   string  `json:"s"`
	BidPrice float64 `json:"b,string"`
	BidQty   float64 `json:"B,string"`
	AskPrice float64 `json:"a,string"`
	AskQty   float64 `json:"A,string"`
}

// NewBinanceExchange creates instance of BinanceExchange
func NewBinanceExchange(pubsub *common.Pubsub) *BinanceExchange {
	keymap := make(map[string]string)
	for _, pair := range config.TradingPairs {
		s1 := pair[0]
		s2 := pair[1]
		if s1 == "USD" {
			s1 = "USDT"
		}
		if s2 == "USD" {
			s2 = "USDT"
		}
		key := fmt.Sprintf("%s%s", s1, s2)
		val := fmt.Sprintf("%s-%s", pair[0], pair[1])
		keymap[key] = val
	}
	ex := &BinanceExchange{"binance", pubsub, keymap}
	return ex
}

// Start begins loop to fetch quotes from Binance
func (ex *BinanceExchange) Start() {
	var streamPaths []string
	for pair := range ex.pubsubKeyMap {
		streamPaths = append(streamPaths, strings.ToLower(pair)+"@bookTicker")
	}
	endpoint := fmt.Sprintf("%s/stream?streams=%s", binanceEndpoint, strings.Join(streamPaths, "/"))
	c, _, err := websocket.DefaultDialer.Dial(endpoint, nil)

	if err != nil {
		log.Fatal("Failed to dial Binance websocket: ", err)
	}
	defer c.Close()

	done := make(chan struct{})

	// Handling incoming messages from websocket
	go func() {
		defer close(done)
		for {
			var payload BinanceWSPayload
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			json.Unmarshal(msg, &payload)
			quote := payload.Data
			pubsubKey := ex.pubsubKeyMap[strings.ToUpper(quote.Symbol)]
			if pubsubKey == "" {
				log.Printf("No pubsub key for %s\n", quote.Symbol)
			}

			quote.AskPrice *= (1.0 + binanceFee)
			quote.BidPrice *= (1.0 - binanceFee)

			ex.pubsub.Publish(
				ex.name,
				pubsubKey,
				&common.ExchangeQuote{Bid: quote.BidPrice, Ask: quote.AskPrice, Exchange: ex.name},
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
