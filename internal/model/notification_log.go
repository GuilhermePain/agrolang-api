package model

import "time"

type NotificationStatus string

const (
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusRead      NotificationStatus = "read"
	NotificationStatusFailed    NotificationStatus = "failed"
)

type NotificationLog struct {
	ID           string             `json:"id"`
	AlertID      *string            `json:"alert_id,omitempty"`
	ProducerID   string             `json:"producer_id"`
	PropertyID   *string            `json:"property_id,omitempty"`
	Phone        string             `json:"phone"`
	Message      string             `json:"message"`
	Status       NotificationStatus `json:"status"`
	SentAt       time.Time          `json:"sent_at"`
	DeliveredAt  *time.Time         `json:"delivered_at,omitempty"`
	ReadAt       *time.Time         `json:"read_at,omitempty"`
	ErrorMessage *string            `json:"error_message,omitempty"`
}

