package output

import (
	"fmt"
	"io"
	"math"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"

	"github.com/ashwinpatri/debtCLI/internal/models"
)

type TableRenderer struct{}

func (r *TableRenderer) Render(w io.Writer, result *models.ScanResult) error {
	if len(result.Snapshot.Items) == 0 {
		fmt.Fprintln(w, "No debt found.")
		return nil
	}

	items := make([]models.DebtItem, len(result.Snapshot.Items))
	copy(items, result.Snapshot.Items)
	sort.Slice(items, func(i, j int) bool {
		if items[i].File != items[j].File {
			return items[i].File < items[j].File
		}
		return items[i].Score > items[j].Score
	})

	tbl := tablewriter.NewWriter(w)
	tbl.Header("FILE", "LINE", "TAG", "AGE", "AUTHOR", "SCORE")

	for _, item := range items {
		age := formatAge(item.Date)
		score := fmt.Sprintf("%.1f", item.Score)
		tag := colorTag(item.Tag)
		if err := tbl.Append(item.File, fmt.Sprintf("%d", item.Line), tag, age, item.Author, score); err != nil {
			return err
		}
	}

	if err := tbl.Render(); err != nil {
		return err
	}

	fmt.Fprintf(w, "\nTotal items:   %d\n", len(items))
	fmt.Fprintf(w, "Health score:  %s\n", formatHealth(result.Snapshot.HealthScore, result.Delta))

	if result.Delta != nil {
		if result.Delta.NewItems > 0 {
			fmt.Fprintf(w, "New debt:      %d item(s)\n", result.Delta.NewItems)
		}
		if result.Delta.ResolvedItems > 0 {
			fmt.Fprintf(w, "Resolved:      %d item(s)\n", result.Delta.ResolvedItems)
		}
	}

	return nil
}

func formatAge(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	days := int(time.Since(t).Hours() / 24)
	switch {
	case days < 1:
		return "today"
	case days == 1:
		return "1d"
	case days < 365:
		return fmt.Sprintf("%dd", days)
	default:
		return fmt.Sprintf("%dy", days/365)
	}
}

func formatHealth(score float64, delta *models.Delta) string {
	scoreStr := fmt.Sprintf("%.0f/100", math.Round(score))
	var colorFn func(string, ...interface{}) string

	switch {
	case score >= 75:
		colorFn = color.GreenString
	case score >= 50:
		colorFn = color.YellowString
	default:
		colorFn = color.RedString
	}

	out := colorFn(scoreStr)
	if delta != nil && delta.ScoreDiff != 0 {
		sign := "▲"
		diff := delta.ScoreDiff
		if diff < 0 {
			sign = "▼"
			diff = -diff
		}
		out += fmt.Sprintf("  (%s %.0f since last scan)", sign, math.Round(diff))
	}
	return out
}

func colorTag(tag string) string {
	switch tag {
	case "SECURITY":
		return color.RedString(tag)
	case "FIXME":
		return color.YellowString(tag)
	case "HACK":
		return color.MagentaString(tag)
	default:
		return tag
	}
}
