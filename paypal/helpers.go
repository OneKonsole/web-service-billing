package paypal

func NewOrderOchestrator() *OrderOrchestrator {
	return &OrderOrchestrator{
		approvalChans: make(map[string]chan bool),
	}
}
