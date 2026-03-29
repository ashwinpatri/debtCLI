// Package output defines the Renderer interface and provides implementations
// for table, JSON, and history chart output formats.
package output

import (
	"io"

	"github.com/ashwinpatri/debtCLI/internal/models"
)

// Renderer writes a scan result to w in a specific format.
// Adding a new output format means implementing this interface and registering
// the implementation in cmd/scan.go — nothing else needs to change.
type Renderer interface {
	Render(w io.Writer, result *models.ScanResult) error
}

// HistoryRenderer writes historical health score data (not a full ScanResult).
type HistoryRenderer interface {
	RenderHistory(w io.Writer, snapshots []*models.Snapshot) error
}
