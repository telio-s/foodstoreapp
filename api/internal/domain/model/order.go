package model

import "time"

type Order struct {
	ID               string
	MemberCardNumber string
	TotalPrice       string
	DiscountAmount   string
	CreatedAt        time.Time
	Items            []*OrderItem
}

type OrderItem struct {
	ID          string
	OrderID     string
	ProductID   string
	ProductName string
	Quantity    int
	UnitPrice   string
}
