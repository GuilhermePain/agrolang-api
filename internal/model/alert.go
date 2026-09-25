package model

import "time"

type Alert struct {
	ID          string
	PropertyID  string
	Level       string
	AlertType   string
	PeriodStart time.Time
	PeriodEnd   time.Time
	TriggeredAt time.Time
	ResolvedAt  *time.Time
}
