// Package scorer computes per-item debt scores and the aggregate repo health score.
// All functions are pure — no side effects, no I/O, no shared state.
package scorer

import (
	"math"
	"time"

	"github.com/ashwinpatri/debtCLI/internal/models"
)

// ScoreItem computes a weighted score for a single debt item.
//
// score = baseSeverity × ageMult × churnMult
//
// ageMult  = 1 + min(ageDays / ageHalfLifeDays, ageMultiplierCap)
// churnMult = 1 + min(churn  / churnSaturationPoint, churnMultiplierCap)
func ScoreItem(item models.DebtItem, baseSeverity float64) float64 {
	ageDays := time.Since(item.Date).Hours() / 24
	ageMult := 1.0 + math.Min(ageDays/ageHalfLifeDays, ageMultiplierCap)
	churnMult := 1.0 + math.Min(float64(item.Churn)/churnSaturationPoint, churnMultiplierCap)
	return baseSeverity * ageMult * churnMult
}

// RepoHealth converts the sum of all item scores into a 0–100 health score.
// A score of 100 means no debt. A score of 0 means total debt at or above baseline.
//
// health = max(0, 100 − (sum / healthScoreBaseline × 100))
func RepoHealth(items []models.DebtItem) float64 {
	var total float64
	for _, item := range items {
		total += item.Score
	}
	health := healthScoreMax - (total/healthScoreBaseline)*healthScoreMax
	return math.Max(0, health)
}
