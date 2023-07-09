package model

import (
	"time"
)

type OrderItemPack struct {
	TableID    *string     `json:"table_id"`
	OrderItems []OrderItem `json:"order_items"`
}

type OrderItemsByOrder struct {
	OrderItems []struct {
		Image      string `json:"image"`
		Name       string `json:"name"`
		Quantity   int    `json:"quantity"`
		TotalPrice int    `json:"total_price"`
		UnitPrice  int    `json:"unit_price"`
	} `json:"order_items"`
	PaymentDue  int `json:"payment_due"`
	TableNumber int `json:"table_number"`
	TotalCount  int `json:"total_count"`
}

type InvoiceViewFormat struct {
	InvoiceID      string
	PaymentMethod  string
	OrderID        string
	PaymentStatus  *string
	OrderDetails   interface{}
	PaymentDue     interface{}
	TableNumber    interface{}
	PaymentDueDate time.Time
}
