package output

import (
	"io"

	"github.com/ashwinpatri/debtCLI/internal/models"
)

type Renderer interface {
	Render(w io.Writer, result *models.ScanResult) error
}

type HistoryRenderer interface {
	RenderHistory(w io.Writer, snapshots []*models.Snapshot) error
}
