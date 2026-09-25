package notifier

import (
	"fmt"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/service/risk"
)

var alertLabels = map[risk.Alert]string{
	risk.AlertSevereWaterStress: "Risco de estresse hídrico severo",
	risk.AlertFrost:             "Risco de geada",
	risk.AlertHeatWave:          "Onda de calor",
	risk.AlertHeavyRain:         "Chuva forte",
}

var alertRecommendations = map[risk.Alert]string{
	risk.AlertSevereWaterStress: "Aumente o turno de irrigação para a noite.",
	risk.AlertFrost:             "Proteja as mudas com cobertura ou antecipe a colheita.",
	risk.AlertHeatWave:          "Evite atividades no campo no horário mais quente do dia.",
	risk.AlertHeavyRain:         "Verifique o escoamento de água e proteja as áreas baixas.",
}

const dateLayout = "02/01"

// BuildMessage renders a plain-language WhatsApp alert (RF-04.2, RNF-04):
// the risk, the event period, and a practical recommendation, with no
// technical weather terms.
func BuildMessage(result risk.Result, periodStart, periodEnd time.Time) string {
	label := alertLabels[result.Alert]
	recommendation := alertRecommendations[result.Alert]

	return fmt.Sprintf(
		"%s\nPeríodo: %s a %s\nRecomendação: %s",
		label,
		periodStart.Format(dateLayout),
		periodEnd.Format(dateLayout),
		recommendation,
	)
}
