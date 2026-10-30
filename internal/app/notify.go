package app

import (
	"slices"

	"github.com/GuilhermePain/agrolang-api/internal/service/risk"
)

// ShouldNotify decides whether a WhatsApp alert goes out. Critical alerts
// always do, because the crop is at stake. Medium alerts are filtered by the
// risk types the account owner chose; an empty list means "notify everything".
func ShouldNotify(level risk.Level, alert risk.Alert, concerns []string) bool {
	if level == risk.LevelCritical {
		return true
	}
	if len(concerns) == 0 {
		return true
	}
	return slices.Contains(concerns, string(alert))
}
