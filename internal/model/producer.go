package model

import "time"

type Producer struct {
	ID            string
	Name          string
	WhatsAppPhone string
	City          string
	State         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
