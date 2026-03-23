package scorer

// Scoring constants that control how age, churn, and health are computed.
// All tuning lives here so that adjustments do not require touching scorer logic.
const (
	// ageHalfLifeDays is the number of days at which the age multiplier
	// reaches 2.0x (one half-life above the baseline of 1.0x).
	ageHalfLifeDays = 180.0

	// ageMultiplierCap is the maximum additive contribution of age before
	// clamping, giving a total multiplier ceiling of 3.0x at roughly one year.
	ageMultiplierCap = 2.0

	// churnSaturationPoint is the commit count at which the churn multiplier
	// reaches its cap. Files touched fewer times than this scale linearly.
	churnSaturationPoint = 50.0

	// churnMultiplierCap is the maximum additive contribution of churn,
	// giving a total multiplier ceiling of 2.0x at saturation point.
	churnMultiplierCap = 1.0

	// healthScoreMax is the best possible repo health score.
	healthScoreMax = 100.0

	// healthScoreBaseline is the total raw debt score that maps to 0 health.
	// A repo whose items sum to this value receives a health score of 0.
	healthScoreBaseline = 200.0
)
