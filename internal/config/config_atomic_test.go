package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// V38: a successful Save leaves no temp file behind — the temp is renamed over
// the target, not left as an orphan.
func TestSave_LeavesNoTempFile_V38(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	res := LoadResult{Config: DefaultConfig()}
	require.NoError(t, Save(path, res))

	matches, err := filepath.Glob(filepath.Join(dir, ".cfgtmp-*"))
	require.NoError(t, err)
	assert.Empty(t, matches, "no temp file after a clean Save")

	// And the written file is loadable (round-trips).
	assert.Equal(t, "tokyo-night", Load(path).Config.Theme)
}

// V38: Load best-effort removes `.cfgtmp-*` files orphaned by a prior crash,
// without disturbing the real config.
func TestLoad_CleansOrphanTemps_V38(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(`theme = "material-dark"`+"\n"), 0o644))

	orphan := filepath.Join(dir, ".cfgtmp-12345")
	require.NoError(t, os.WriteFile(orphan, []byte("garbage"), 0o644))

	res := Load(path)
	assert.Equal(t, "material-dark", res.Config.Theme, "real config still loaded")

	_, statErr := os.Stat(orphan)
	assert.True(t, os.IsNotExist(statErr), "orphan temp removed on Load")
}

// V38: a crash between marshal and rename leaves the original file byte-for-byte
// intact (never partial). The crash is simulated in a subprocess that exits 99
// at the rename instrument; the parent verifies the file is unchanged.
func TestAtomicSave_CrashBeforeRename_V38(t *testing.T) {
	if os.Getenv("GLOGGY_RUN_CRASH_SAVE") == "1" {
		// Subprocess: perform a Save that exits before the rename.
		path := os.Getenv("GLOGGY_CRASH_PATH")
		res := Load(path)
		res.Config.Theme = "solarized-dark"
		_ = Save(path, res) // os.Exit(99) fires inside atomicWrite
		return              // unreachable
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	preWrite := `theme = "catppuccin-mocha"` + "\n" + `keep_me = "yes"` + "\n"
	require.NoError(t, os.WriteFile(path, []byte(preWrite), 0o644))

	cmd := exec.Command(os.Args[0], "-test.run=^TestAtomicSave_CrashBeforeRename_V38$")
	cmd.Env = append(os.Environ(),
		"GLOGGY_RUN_CRASH_SAVE=1",
		crashEnvBeforeRename+"=1",
		"GLOGGY_CRASH_PATH="+path,
	)
	err := cmd.Run()

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr, "subprocess must exit non-zero (crashed)")
	assert.Equal(t, 99, exitErr.ExitCode(), "exited at the pre-rename instrument")

	// The original file is byte-for-byte intact — never partial (V38).
	after, rerr := os.ReadFile(path)
	require.NoError(t, rerr)
	assert.Equal(t, preWrite, string(after), "config unchanged by the crashed Save")

	// Load recovers cleanly and sweeps the orphan temp.
	res := Load(path)
	assert.Equal(t, "catppuccin-mocha", res.Config.Theme)
	matches, _ := filepath.Glob(filepath.Join(dir, ".cfgtmp-*"))
	assert.Empty(t, matches, "orphan temp cleaned on Load")
}
