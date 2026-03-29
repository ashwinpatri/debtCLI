package output

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/ashwinpatri/debtCLI/internal/models"
)

const (
	barMaxWidth = 40
	barChar     = "█"
)

// HistoryTableRenderer writes an ASCII bar chart of health scores over time.
type HistoryTableRenderer struct{}

// RenderHistory writes a bar chart of the health score for each snapshot to w.
func (r *HistoryTableRenderer) RenderHistory(w io.Writer, snapshots []*models.Snapshot) error {
	if len(snapshots) == 0 {
		fmt.Fprintln(w, "No history found for this repository.")
		return nil
	}

	fmt.Fprintln(w, "HEALTH SCORE HISTORY")
	fmt.Fprintln(w, strings.Repeat("─", 60))

	for _, snap := range snapshots {
		date := snap.Timestamp.Format("2006-01-02 15:04")
		score := math.Round(snap.HealthScore)
		barLen := int(score * barMaxWidth / 100)
		bar := strings.Repeat(barChar, barLen)

		fmt.Fprintf(w, "%s  %s%-*s  %3.0f/100  (%d items)\n",
			date, bar, barMaxWidth-barLen, "", score, snap.ItemCount)
	}

	return nil
}
