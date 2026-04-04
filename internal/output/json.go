package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ashwinpatri/debtCLI/internal/models"
)

type JSONRenderer struct{}

func (r *JSONRenderer) Render(w io.Writer, result *models.ScanResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	return nil
}
