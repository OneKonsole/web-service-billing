package paypal

import (
	"sync"

	oko "github.com/OneKonsole/order-model"
)

// Paypal related
type PaypalOrderInfos struct {
	Order          oko.Order `json:"order_details"`
	CurrencyCode   string    `json:"currency"`
	MaxAmountValue string    `json:"amount"`
}
type PaypalOrderResponse struct {
	OrderID string            `json:"id"`
	Status  string            `json:"status"`
	Links   []PaypalOrderLink `json:"links"`
}

type PaypalOrderLink struct {
	Href   string `json:"href"`
	Rel    string `json:"rel"`
	Method string `json:"method"`
}

// Generic order related
type OrderOrchestrator struct {
	approvalChans map[string]chan bool
	mutex         sync.Mutex
}
