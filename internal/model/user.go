package model

import "time"

type Profile string

const (
	ProfileProducer   Profile = "producer"
	ProfileTechnician Profile = "technician"
)

type User struct {
	ID            string            `json:"id"`
	Email         string            `json:"email"`
	PasswordHash  string            `json:"-"`
	Profile       Profile           `json:"profile"`
	Name          string            `json:"name"`
	WhatsAppPhone string            `json:"whatsapp_phone"`
	Organization  *string           `json:"organization"`
	Concerns      []string          `json:"concerns"`
	Onboarding    map[string]string `json:"onboarding"`
	CreatedAt     time.Time         `json:"created_at"`
}
