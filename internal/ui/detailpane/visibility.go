package detailpane

import (
	"slices"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
)

// VisibilityModel tracks which fields are hidden and persists changes to config.
type VisibilityModel struct {
	hiddenFields []string
	configPath   string
	loadResult   config.LoadResult
}

// NewVisibilityModel creates a VisibilityModel from an existing load result.
func NewVisibilityModel(configPath string, lr config.LoadResult) VisibilityModel {
	hidden := make([]string, len(lr.Config.HiddenFields))
	copy(hidden, lr.Config.HiddenFields)
	return VisibilityModel{
		hiddenFields: hidden,
		configPath:   configPath,
		loadResult:   lr,
	}
}

// HiddenFields returns the current set of hidden field names.
func (m VisibilityModel) HiddenFields() []string {
	out := make([]string, len(m.hiddenFields))
	copy(out, m.hiddenFields)
	return out
}

// ToggleField hides a visible field or shows a hidden field, then writes to config.
// Returns updated model and any write error.
func (m VisibilityModel) ToggleField(field string) (VisibilityModel, error) {
	idx := slices.Index(m.hiddenFields, field)
	if idx >= 0 {
		// Currently hidden → show it.
		m.hiddenFields = slices.Delete(slices.Clone(m.hiddenFields), idx, idx+1)
	} else {
		// Currently visible → hide it.
		m.hiddenFields = append(slices.Clone(m.hiddenFields), field)
	}
	m.loadResult.Config.HiddenFields = m.hiddenFields
	if m.configPath != "" {
		if err := config.Save(m.configPath, m.loadResult); err != nil {
			return m, err
		}
	}
	return m, nil
}

// RenderEntry renders an entry using the current hidden fields.
// Returns syntax-highlighted JSON for JSONL entries, plain text otherwise.
func (m VisibilityModel) RenderEntry(entry logsource.Entry, th theme.Theme) string {
	if entry.IsJSON {
		return RenderJSON(entry, th, m.hiddenFields)
	}
	return RenderRaw(entry)
}
