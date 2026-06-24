package external

import (
	"errors"
	"sync/atomic"
)

var ErrThirdPartyFailure = errors.New("third-party accounting service mengembalikan 500 internal server error")

type ThirdPartyClient interface {
	SendTransaction(payload Payload) error
}

type Payload struct {
	TransactionID string
	Amount        float64
}

type MockFlakyClient struct {
	FailFirstNCalls int32
	callCount       int32
}

func (c *MockFlakyClient) SendTransaction(payload Payload) error {
	n := atomic.AddInt32(&c.callCount, 1)
	if n <= c.FailFirstNCalls {
		return ErrThirdPartyFailure
	}
	return nil
}
