package helpers

import (
	b64 "encoding/base64"
	"fmt"

	oko "github.com/OneKonsole/order-model"
)

type defaultPrices struct {
	Basic             int `json:"basic"`
	ImageStorage      int `json:"img_storage_price_unit"`
	MonitoringOption  int `json:"monitoring_option"`
	MonitoringStorage int `json:"monitoring_storage_price_unit"`
	AlertingOption    int `json:"alerting_option"`
}

func NewPrices() *defaultPrices {
	return &defaultPrices{
		Basic:             20,
		ImageStorage:      1,
		MonitoringOption:  5,
		MonitoringStorage: 1,
		AlertingOption:    5,
	}
}
func AggregateClientInformation(clientID string, clientSecret string) string {
	aggregatedInfos := fmt.Sprintf("%s:%s", clientID, clientSecret)
	encodedInfos := b64.StdEncoding.EncodeToString([]byte(aggregatedInfos))
	return encodedInfos
}

func CalculatePrice(order *oko.Order) int {
	defaultPrices := NewPrices()

	sum := defaultPrices.Basic
	sum += defaultPrices.ImageStorage * order.ImageStorage
	if order.HasMonitoring {
		sum += defaultPrices.MonitoringOption
		sum += defaultPrices.MonitoringStorage * order.MonitoringStorage
	}
	if order.HasAlerting {
		sum += defaultPrices.AlertingOption
	}

	return sum
}
