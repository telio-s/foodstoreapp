package model

import "time"

type Order struct {
	ID               int64
	MemberCardNumber string
	TotalPrice       float64
	DiscountAmount   float64
	CreatedAt        time.Time
	Items            []*OrderItem
}

type OrderItem struct {
	ID        int64
	OrderID   int64
	ProductID int64
	Quantity  int
	UnitPrice float64
}
