package dashboard

import (
	"geocash/internal/domain/terminal"
	"geocash/internal/platform/ai"
)

type DashboardResponse struct {
	Forte           []terminal.ATM      `json:"forte"`
	Competitors     []terminal.ATM      `json:"competitors"`
	HeatmapGrid     interface{}         `json:"heatmapGrid"`
	Recommendations []ai.Recommendation `json:"recommendations"` // <-- Добавить это
}
