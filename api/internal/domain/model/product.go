package model

import "time"

type Product struct {
	ID          string
	Name        string
	Price       string
	IsLimited   bool
	LastOrderAt *time.Time
}
