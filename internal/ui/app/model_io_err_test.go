package app

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/logsource"
)

// V37d/V43: a LoadErrMsg finishes loading and surfaces a visible notice rather
// than silently completing as an empty load. Covers both the direct case and
// the LoadFileStreamMsg-wrapped case used by the live stream.
func TestModel_LoadErr_SurfacesNotice_V37(t *testing.T) {
	loadErr := errors.New("permission denied")

	t.Run("direct", func(t *testing.T) {
		m := resize(newModel(), 120, 30)
		m = send(m, logsource.LoadErrMsg{Err: loadErr})
		assert.False(t, m.loading.IsActive(), "loading marked done on error")
		require.True(t, m.keyhints.HasNotice(), "notice must be set")
		assert.Contains(t, m.View(), "load error", "error surfaced to user")
		assert.Contains(t, m.View(), "permission denied")
	})

	t.Run("wrapped in stream", func(t *testing.T) {
		m := resize(newModel(), 120, 30)
		m = send(m, logsource.NewLoadFileStreamMsgForTest(logsource.LoadErrMsg{Err: loadErr}))
		assert.False(t, m.loading.IsActive())
		assert.True(t, m.keyhints.HasNotice())
		assert.Contains(t, m.View(), "load error")
	})
}

// V37c/V43/V3: a retryable TailErrMsg surfaces a notice, keeps draining the
// stream (non-nil cmd → the backing-off goroutine continues), and does NOT drop
// follow mode ([FOLLOW] stays live).
func TestModel_TailErr_Retryable_NoticeAndKeepsDraining_V37(t *testing.T) {
	m := resize(newModel(), 120, 30)
	m.followMode = true

	updated, cmd := m.Update(logsource.NewTailStreamMsgForTest(
		logsource.TailErrMsg{Err: errors.New("input/output error"), Retryable: true}))
	m = updated.(Model)

	require.True(t, m.keyhints.HasNotice())
	view := m.View()
	assert.Contains(t, view, "tail error")
	assert.Contains(t, view, "retrying")
	assert.True(t, m.followMode, "retryable tail error must not drop follow (V3)")
	assert.NotNil(t, cmd, "must keep draining the tail stream")
}

// V37c: a non-retryable TailErrMsg phrases the notice as stopped.
func TestModel_TailErr_NonRetryable_StoppedNotice_V37(t *testing.T) {
	m := resize(newModel(), 120, 30)
	m = send(m, logsource.NewTailStreamMsgForTest(
		logsource.TailErrMsg{Err: errors.New("device gone"), Retryable: false}))

	require.True(t, m.keyhints.HasNotice())
	assert.Contains(t, m.View(), "tail stopped")
	assert.False(t, strings.Contains(m.View(), "retrying"), "non-retryable is not phrased as retrying")
}
