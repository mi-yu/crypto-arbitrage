package common

import (
	"fmt"
)

// ExchangeQuote represents a quote from an exchange
type ExchangeQuote struct {
	Exchange string
	Bid      float64
	Ask      float64
}

func (quote *ExchangeQuote) String() string {
	return fmt.Sprintf("%s - (Bid: %.6f, Ask: %.6f)", quote.Exchange, quote.Bid, quote.Ask)
}

// Exchange defines the API that all exchange nodes should provide
type Exchange interface {
	// Start fetching quotes
	Start()
}
