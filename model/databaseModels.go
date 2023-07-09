package model

import (
	"time"
)

// food model
type Food struct {
	FoodID    string    `json:"_key"`
	Name      *string   `json:"name" validate:"required,min=3,max=30"`
	UnitPrice *float64  `json:"unit_price" validate:"required"`
	FoodImage *string   `json:"food_image" validate:"required"`
	MenuID    *string   `json:"menu_id" validate:"required"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// order model
type Order struct {
	OrderID   string    `json:"_key"`
	TableID   *string   `json:"table_id" validate:"required"`
	OrderDate time.Time `json:"order_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// orderItem model
type OrderItem struct {
	OrderItemID string    `json:"_key"`
	FoodID      *string   `json:"food_id" validate:"required"`
	Quantity    *float64  `json:"quantity" validate:"required"`
	TotalPrice  *float64  `json:"total_price"`
	OrderID     string    `json:"order_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// invoice model
type Invoice struct {
	InvoiceID      string    `json:"_key"`
	OrderID        string    `json:"order_id" validate:"required"`
	PaymentMethod  *string   `json:"payment_method" validate:"eq=CARD|eq=CASH|eq="`
	PaymentStatus  *string   `json:"payment_status" validate:"required,eq=PENDING|eq=PAID"`
	PaymentDueDate time.Time `json:"payment_due_date"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// menu model
type Menu struct {
	MenuID    string     `json:"_key"`
	Name      string     `json:"name" validate:"required"`
	Category  string     `json:"category" validate:"required"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// table model
type Table struct {
	TableID       string    `json:"_key"`
	NumberOfGuest *int      `json:"number_of_guest" validate:"required"`
	TableNumber   *int      `json:"table_number" validate:"required"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
