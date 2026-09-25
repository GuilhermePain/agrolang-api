package model

import "time"

type Alert struct {
	ID          string     `json:"id"`
	PropertyID  string     `json:"property_id"`
	Level       string     `json:"level"`
	AlertType   string     `json:"alert_type"`
	PeriodStart time.Time  `json:"period_start"`
	PeriodEnd   time.Time  `json:"period_end"`
	TriggeredAt time.Time  `json:"triggered_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
}
