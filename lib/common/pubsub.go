package common

import (
	"sync"

	"github.com/mi-yu/crypto-arbitrage/lib/config"
)

// Pubsub object
type Pubsub struct {
	lock sync.RWMutex
	// Keyed by <exchange_name>, then <trading_pair_key>
	subs   map[string]map[string][]chan *ExchangeQuote
	closed bool
}

// NewPubsub creates a new Pubsub instance
func NewPubsub() *Pubsub {
	ps := &Pubsub{}
	ps.subs = make(map[string]map[string][]chan *ExchangeQuote)
	for _, ex := range config.Exchanges {
		ps.subs[ex] = make(map[string][]chan *ExchangeQuote)
	}
	return ps
}

// Subscribe takes a topic and channel, subscribing the channel
// to any pubs from the topic
func (ps *Pubsub) Subscribe(exchange string, pairString string, ch chan *ExchangeQuote) {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if ps.closed {
		return
	}

	ps.subs[exchange][pairString] = append(ps.subs[exchange][pairString], ch)
}

// Publish takes a topic and message, sending message to
// all channels subscribed to said topic
func (ps *Pubsub) Publish(exchange string, pairString string, msg *ExchangeQuote) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	if ps.closed {
		return
	}

	for _, ch := range ps.subs[exchange][pairString] {
		ch <- msg
	}
}

// Close cleans up all channels
func (ps *Pubsub) Close() {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if !ps.closed {
		ps.closed = true
		for _, exchanges := range ps.subs {
			for _, pair := range exchanges {
				for _, ch := range pair {
					close(ch)
				}
			}
		}
	}
}
