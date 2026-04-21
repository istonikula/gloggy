package entrylist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbbreviateLogger(t *testing.T) {
	tests := []struct {
		name     string
		logger   string
		depth    int
		expected string
	}{
		{
			// R2 criterion 1
			name:     "depth2 5segments",
			logger:   "org.springframework.data.repository.RepositoryDelegate",
			depth:    2,
			expected: "o.s.d.repository.RepositoryDelegate",
		},
		{
			// R2 criterion 2: keep last 2 full, abbreviate first 2
			name:     "depth2 4segments",
			logger:   "com.example.server.AppServerKt",
			depth:    2,
			expected: "c.e.server.AppServerKt",
		},
		{
			// R2 criterion 3
			name:     "depth1 4segments",
			logger:   "com.example.server.AppServerKt",
			depth:    1,
			expected: "c.e.s.AppServerKt",
		},
		{
			// R2 criterion 4: fewer segments than depth → unchanged
			name:     "fewer segments than depth",
			logger:   "AppServerKt",
			depth:    2,
			expected: "AppServerKt",
		},
		{
			name:     "exactly depth segments unchanged",
			logger:   "server.AppServerKt",
			depth:    2,
			expected: "server.AppServerKt",
		},
		{
			name:     "single segment",
			logger:   "Main",
			depth:    1,
			expected: "Main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AbbreviateLogger(tt.logger, tt.depth)
			assert.Equal(t, tt.expected, got, "AbbreviateLogger(%q, %d)", tt.logger, tt.depth)
		})
	}
}
