package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(file))
}

func runCmd(t *testing.T, home string, args ...string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = repoRoot()
	cmd.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home)
	return cmd
}

func mustRun(t *testing.T, home string, args ...string) (stdout, stderr string) {
	t.Helper()
	cmd := runCmd(t, home, args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("command %v failed: %v\nstderr: %s", args, err, ee.Stderr)
		}
		t.Fatalf("command %v failed: %v", args, err)
	}
	se := ""
	if ee, ok := err.(*exec.ExitError); ok {
		se = string(ee.Stderr)
	}
	return string(out), se
}

func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}
	var v map[string]any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("cannot parse %s: %v", path, err)
	}
	return v
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// TestInstallUpgradesOldTwoFileInstall verifies that running install.sh over an
// old two-file install (no statusline) adds caveman-statusline.sh and wires it.
func TestInstallUpgradesOldTwoFileInstall(t *testing.T) {
	home := t.TempDir()
	hooksDir := filepath.Join(home, ".claude", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(home, ".claude", "settings.json"), map[string]any{})
	_ = os.WriteFile(filepath.Join(hooksDir, "caveman-activate.js"), nil, 0644)
	_ = os.WriteFile(filepath.Join(hooksDir, "caveman-mode-tracker.js"), nil, 0644)

	mustRun(t, home, "bash", "hooks/install.sh")

	statusline := filepath.Join(hooksDir, "caveman-statusline.sh")
	if _, err := os.Stat(statusline); err != nil {
		t.Error("upgrade should install caveman-statusline.sh")
	}

	settings := readJSON(t, filepath.Join(home, ".claude", "settings.json"))
	sl, ok := settings["statusLine"].(map[string]any)
	if !ok {
		t.Fatal("settings.json missing statusLine after upgrade")
	}
	cmd, _ := sl["command"].(string)
	if cmd == "" || !contains(cmd, statusline) {
		t.Errorf("statusLine.command should reference %s, got %q", statusline, cmd)
	}
}

// TestInstallReconfiguresMissingStatusline verifies that install.sh adds the
// statusLine config when hooks are already installed but statusLine is absent.
func TestInstallReconfiguresMissingStatusline(t *testing.T) {
	home := t.TempDir()
	claudeDir := filepath.Join(home, ".claude")
	hooksDir := filepath.Join(claudeDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"caveman-activate.js", "caveman-mode-tracker.js", "caveman-statusline.sh"} {
		_ = os.WriteFile(filepath.Join(hooksDir, name), nil, 0644)
	}
	settings := map[string]any{
		"hooks": map[string]any{
			"SessionStart": []any{
				map[string]any{"hooks": []any{map[string]any{
					"type":    "command",
					"command": `node "` + filepath.Join(hooksDir, "caveman-activate.js") + `"`,
				}}},
			},
			"UserPromptSubmit": []any{
				map[string]any{"hooks": []any{map[string]any{
					"type":    "command",
					"command": `node "` + filepath.Join(hooksDir, "caveman-mode-tracker.js") + `"`,
				}}},
			},
		},
	}
	writeJSON(t, filepath.Join(claudeDir, "settings.json"), settings)

	stdout, _ := mustRun(t, home, "bash", "hooks/install.sh")
	if contains(stdout, "Nothing to do") {
		t.Error("install.sh should not report 'Nothing to do' when statusLine is missing")
	}

	updated := readJSON(t, filepath.Join(claudeDir, "settings.json"))
	sl, ok := updated["statusLine"].(map[string]any)
	if !ok {
		t.Fatal("settings.json missing statusLine after reconfigure")
	}
	cmd, _ := sl["command"].(string)
	if !contains(cmd, filepath.Join(hooksDir, "caveman-statusline.sh")) {
		t.Errorf("statusLine.command missing statusline path, got %q", cmd)
	}
}

// TestUninstallPreservesCustomStatusline verifies that uninstall.sh removes hooks
// but does not overwrite a custom (non-caveman) statusLine command.
func TestUninstallPreservesCustomStatusline(t *testing.T) {
	home := t.TempDir()
	claudeDir := filepath.Join(home, ".claude")
	hooksDir := filepath.Join(claudeDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"caveman-activate.js", "caveman-mode-tracker.js", "caveman-statusline.sh"} {
		_ = os.WriteFile(filepath.Join(hooksDir, name), nil, 0644)
	}
	customCmd := "bash /tmp/custom-status-with-caveman.sh"
	settings := map[string]any{
		"statusLine": map[string]any{"type": "command", "command": customCmd},
		"hooks": map[string]any{
			"SessionStart": []any{map[string]any{"hooks": []any{map[string]any{
				"type": "command", "command": `node "` + filepath.Join(hooksDir, "caveman-activate.js") + `"`,
			}}}},
			"UserPromptSubmit": []any{map[string]any{"hooks": []any{map[string]any{
				"type": "command", "command": `node "` + filepath.Join(hooksDir, "caveman-mode-tracker.js") + `"`,
			}}}},
		},
	}
	writeJSON(t, filepath.Join(claudeDir, "settings.json"), settings)

	mustRun(t, home, "bash", "hooks/uninstall.sh")

	updated := readJSON(t, filepath.Join(claudeDir, "settings.json"))
	sl, ok := updated["statusLine"].(map[string]any)
	if !ok {
		t.Fatal("statusLine removed after uninstall — should be preserved")
	}
	if sl["command"] != customCmd {
		t.Errorf("statusLine.command changed: want %q, got %q", customCmd, sl["command"])
	}
	if _, hasHooks := updated["hooks"]; hasHooks {
		t.Error("hooks key should be gone after uninstall")
	}
}

// TestActivateDoesNotNudgeWhenCustomStatuslineExists verifies that
// caveman-activate.js does not print a statusline setup nudge when a custom
// statusLine already exists, and that it writes the flag file.
func TestActivateDoesNotNudgeWhenCustomStatuslineExists(t *testing.T) {
	home := t.TempDir()
	claudeDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(claudeDir, "settings.json"), map[string]any{
		"statusLine": map[string]any{"type": "command", "command": "bash /tmp/my-statusline.sh"},
	})

	stdout, _ := mustRun(t, home, "node", "hooks/caveman-activate.js")

	if contains(stdout, "STATUSLINE SETUP NEEDED") {
		t.Error("should not nudge for statusline when custom one already exists")
	}
	flag := filepath.Join(claudeDir, ".caveman-active")
	data, err := os.ReadFile(flag)
	if err != nil {
		t.Fatalf("flag file not written: %v", err)
	}
	if string(data) != "full" {
		t.Errorf("flag file: want %q, got %q", "full", string(data))
	}
}

func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || len(s) > 0 && indexStr(s, substr) >= 0)
}

func indexStr(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
