package scorer

import (
	"testing"
	"time"

	"github.com/ashwinpatri/debtCLI/internal/models"
)

func TestScoreItem_ZeroAge(t *testing.T) {
	// An item committed right now: ageDays ≈ 0, ageMult ≈ 1.0, churnMult = 1.0.
	item := models.DebtItem{
		Date:  time.Now(),
		Churn: 0,
	}
	got := ScoreItem(item, 2.0)
	// ageMult ≈ 1.0, churnMult = 1.0 → score ≈ 2.0
	if got < 1.9 || got > 2.2 {
		t.Errorf("score for zero-age item: got %.4f, want ≈2.0", got)
	}
}

func TestScoreItem_OldHighChurn(t *testing.T) {
	// Item 360 days old (2× half-life) and at churn saturation.
	item := models.DebtItem{
		Date:  time.Now().AddDate(0, 0, -360),
		Churn: 50,
	}
	got := ScoreItem(item, 3.0)
	// ageMult = 1 + min(360/180, 2) = 1 + 2 = 3.0
	// churnMult = 1 + min(50/50, 1) = 1 + 1 = 2.0
	// score = 3.0 * 3.0 * 2.0 = 18.0
	if got < 17.5 || got > 18.5 {
		t.Errorf("score for old high-churn item: got %.4f, want ≈18.0", got)
	}
}

func TestScoreItem_AgeCap(t *testing.T) {
	// Age beyond 2 half-lives should be capped at ageMultiplierCap.
	old := models.DebtItem{Date: time.Now().AddDate(-5, 0, 0), Churn: 0}
	veryOld := models.DebtItem{Date: time.Now().AddDate(-10, 0, 0), Churn: 0}

	scoreOld := ScoreItem(old, 1.0)
	scoreVeryOld := ScoreItem(veryOld, 1.0)

	if scoreOld != scoreVeryOld {
		// Both should be capped at the same value.
		t.Errorf("age cap not applied: old=%.4f veryOld=%.4f", scoreOld, scoreVeryOld)
	}
}

func TestScoreItem_ChurnCap(t *testing.T) {
	item50 := models.DebtItem{Date: time.Now(), Churn: 50}
	item200 := models.DebtItem{Date: time.Now(), Churn: 200}

	score50 := ScoreItem(item50, 1.0)
	score200 := ScoreItem(item200, 1.0)

	if score50 != score200 {
		t.Errorf("churn cap not applied: churn50=%.4f churn200=%.4f", score50, score200)
	}
}

func TestRepoHealth_Clean(t *testing.T) {
	health := RepoHealth(nil)
	if health != 100.0 {
		t.Errorf("empty repo health: got %.1f, want 100.0", health)
	}
}

func TestRepoHealth_OnFire(t *testing.T) {
	// Total score equals the baseline — health should be 0.
	items := make([]models.DebtItem, 1)
	items[0].Score = healthScoreBaseline
	health := RepoHealth(items)
	if health != 0.0 {
		t.Errorf("baseline-score health: got %.1f, want 0.0", health)
	}
}

func TestRepoHealth_NeverNegative(t *testing.T) {
	items := []models.DebtItem{{Score: healthScoreBaseline * 10}}
	health := RepoHealth(items)
	if health < 0 {
		t.Errorf("health went negative: %.1f", health)
	}
}

func TestRepoHealth_Partial(t *testing.T) {
	// Half the baseline → health = 50.
	items := []models.DebtItem{{Score: healthScoreBaseline / 2}}
	health := RepoHealth(items)
	if health < 49.9 || health > 50.1 {
		t.Errorf("half-baseline health: got %.1f, want 50.0", health)
	}
}
