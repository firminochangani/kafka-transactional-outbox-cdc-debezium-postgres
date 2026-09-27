package events

import "time"

const (
	TopicOrders = "work.firmino.orders"
)

type EventHeader struct {
	CorrelationID string `json:"correlation_id"`
	Service       string `json:"service"`
	Topic         string `json:"topic"`
}

type OrderPlaced struct {
	EventHeader EventHeader       `json:"header"`
	OrderID     string            `json:"order_id"`
	Items       []OrderPlacedItem `json:"items"`
	PlacedAt    time.Time         `json:"placed_at"`
}

type OrderPlacedItem struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Price     int    `json:"price"`
	Quantity  int    `json:"quantity"`
}
