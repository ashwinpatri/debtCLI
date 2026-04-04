package scorer

import (
	"math"
	"time"

	"github.com/ashwinpatri/debtCLI/internal/models"
)

// ScoreItem computes: severity × (1 + min(ageDays/180, 2)) × (1 + min(churn/50, 1))
func ScoreItem(item models.DebtItem, baseSeverity float64) float64 {
	ageDays := time.Since(item.Date).Hours() / 24
	ageMult := 1.0 + math.Min(ageDays/ageHalfLifeDays, ageMultiplierCap)
	churnMult := 1.0 + math.Min(float64(item.Churn)/churnSaturationPoint, churnMultiplierCap)
	return baseSeverity * ageMult * churnMult
}

// RepoHealth computes: max(0, 100 − (sum(scores) / 200 × 100))
func RepoHealth(items []models.DebtItem) float64 {
	var total float64
	for _, item := range items {
		total += item.Score
	}
	return math.Max(0, healthScoreMax-(total/healthScoreBaseline)*healthScoreMax)
}
