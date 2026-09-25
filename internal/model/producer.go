package model

import "time"

type Producer struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	WhatsAppPhone string    `json:"whatsapp_phone"`
	City          string    `json:"city"`
	State         string    `json:"state"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
