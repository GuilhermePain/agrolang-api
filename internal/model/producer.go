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

	// Concerns are the risk types the producer's account owner asked to be
	// alerted about; empty means "notify everything". Filled by GetByID only
	// and never serialized.
	Concerns []string `json:"-"`
}
