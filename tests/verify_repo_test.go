package integration

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func root() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(file))
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}
	return string(data)
}

func runVerify(t *testing.T, env map[string]string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = root()
	e := os.Environ()
	for k, v := range env {
		e = append(e, k+"="+v)
	}
	cmd.Env = e
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			code = 1
		}
	}
	return stdout.String(), stderr.String(), code
}

// TestSyncedFiles checks that SKILL.md copies match the source and that
// the iceage.skill ZIP is valid and contains the right payload.
func TestSyncedFiles(t *testing.T) {
	r := root()
	skillSource := readFile(t, filepath.Join(r, "skills/iceage/SKILL.md"))

	copies := []string{
		filepath.Join(r, "plugins/iceage/skills/iceage/SKILL.md"),
	}
	for _, c := range copies {
		if _, err := os.Stat(c); os.IsNotExist(err) {
			t.Skipf("synced copy not present (CI may not have run yet): %s", c)
		}
		got := readFile(t, c)
		if got != skillSource {
			t.Errorf("skill copy mismatch: %s", c)
		}
	}

	skillZip := filepath.Join(r, "iceage.skill")
	if _, err := os.Stat(skillZip); os.IsNotExist(err) {
		t.Skip("iceage.skill not present (CI may not have run yet)")
	}
	zr, err := zip.OpenReader(skillZip)
	if err != nil {
		t.Fatalf("cannot open iceage.skill: %v", err)
	}
	defer zr.Close()

	found := false
	for _, f := range zr.File {
		if f.Name == "iceage/SKILL.md" {
			found = true
			rc, _ := f.Open()
			data := make([]byte, f.UncompressedSize64)
			rc.Read(data)
			rc.Close()
			if string(data) != skillSource {
				t.Error("iceage.skill payload mismatch")
			}
		}
	}
	if !found {
		t.Error("iceage.skill missing iceage/SKILL.md")
	}
}

// TestManifestsAndSyntax validates JSON manifests and JS/bash script syntax.
func TestManifestsAndSyntax(t *testing.T) {
	r := root()

	manifestPaths := []string{
		filepath.Join(r, ".agents/plugins/marketplace.json"),
		filepath.Join(r, ".claude-plugin/plugin.json"),
		filepath.Join(r, ".claude-plugin/marketplace.json"),
		filepath.Join(r, ".codex/hooks.json"),
		filepath.Join(r, "gemini-extension.json"),
		filepath.Join(r, "plugins/iceage/.codex-plugin/plugin.json"),
	}
	for _, p := range manifestPaths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Logf("skipping missing manifest: %s", p)
			continue
		}
		data, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("cannot read %s: %v", p, err)
			continue
		}
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			t.Errorf("invalid JSON in %s: %v", p, err)
		}
	}

	jsFiles := []string{"hooks/caveman-config.js", "hooks/caveman-activate.js", "hooks/caveman-mode-tracker.js"}
	for _, f := range jsFiles {
		stdout, stderr, code := runVerify(t, nil, "node", "--check", f)
		if code != 0 {
			t.Errorf("node --check %s failed (code %d):\n%s\n%s", f, code, stdout, stderr)
		}
	}

	bashFiles := []string{"hooks/install.sh", "hooks/uninstall.sh", "hooks/caveman-statusline.sh"}
	for _, f := range bashFiles {
		stdout, stderr, code := runVerify(t, nil, "bash", "-n", f)
		if code != 0 {
			t.Errorf("bash -n %s failed (code %d):\n%s\n%s", f, code, stdout, stderr)
		}
	}

	installSh := readFile(t, filepath.Join(r, "hooks/install.sh"))
	uninstallSh := readFile(t, filepath.Join(r, "hooks/uninstall.sh"))
	if !strings.Contains(installSh, "caveman-config.js") {
		t.Error("install.sh missing caveman-config.js")
	}
	if !strings.Contains(uninstallSh, "caveman-config.js") {
		t.Error("uninstall.sh missing caveman-config.js")
	}
}

// TestPowerShellStatic checks static PS 5.1 compatibility in PowerShell scripts.
func TestPowerShellStatic(t *testing.T) {
	r := root()
	installText := readFile(t, filepath.Join(r, "hooks/install.ps1"))
	uninstallText := readFile(t, filepath.Join(r, "hooks/uninstall.ps1"))
	statuslineText := readFile(t, filepath.Join(r, "hooks/caveman-statusline.ps1"))

	checks := []struct{ text, substr, msg string }{
		{installText, "caveman-config.js", "install.ps1 missing caveman-config.js"},
		{uninstallText, "caveman-config.js", "uninstall.ps1 missing caveman-config.js"},
		{installText, "caveman-statusline.ps1", "install.ps1 missing statusline.ps1"},
		{uninstallText, "caveman-statusline.ps1", "uninstall.ps1 missing statusline.ps1"},
		{statuslineText, "[CAVEMAN", "caveman-statusline.ps1 missing badge output"},
		{installText, "powershell -ExecutionPolicy Bypass -File",
			"install.ps1 missing PowerShell statusline command"},
	}
	for _, c := range checks {
		if !strings.Contains(c.text, c.substr) {
			t.Error(c.msg)
		}
	}
	if strings.Contains(installText, "-AsHashtable") {
		t.Error("install.ps1 should stay compatible with Windows PowerShell 5.1 (no -AsHashtable)")
	}
}

// TestCompressFixtures validates that pre-existing compressed fixture pairs
// satisfy the basic structural invariants (headings, code blocks, URLs).
func TestCompressFixtures(t *testing.T) {
	r := root()
	fixtureDir := filepath.Join(r, "tests/caveman-compress")
	entries, err := filepath.Glob(filepath.Join(fixtureDir, "*.original.md"))
	if err != nil || len(entries) == 0 {
		t.Skip("no caveman-compress fixtures found")
	}

	for _, original := range entries {
		name := filepath.Base(original)
		compressed := filepath.Join(fixtureDir, strings.Replace(name, ".original.md", ".md", 1))
		if _, err := os.Stat(compressed); os.IsNotExist(err) {
			t.Errorf("missing compressed fixture for %s", name)
			continue
		}
		origText := readFile(t, original)
		compText := readFile(t, compressed)

		if errs := validateCompressed(origText, compText); len(errs) > 0 {
			for _, e := range errs {
				t.Errorf("fixture %s: %s", filepath.Base(compressed), e)
			}
		}

		// File should be detectable as natural language (not code/binary)
		if !isNaturalLanguage(compressed) {
			t.Errorf("fixture %s should be natural language", filepath.Base(compressed))
		}
	}
}

// TestCompressCLI verifies the iceage-compress CLI skip path (code file → exit 0)
// and the missing-file error path (exit 1).
func TestCompressCLI(t *testing.T) {
	r := root()
	goDir := filepath.Join(r, "iceage-compress", "go")

	if _, err := os.Stat(goDir); os.IsNotExist(err) {
		t.Skip("iceage-compress/go not present")
	}

	// Skip path: code file → detected as code, exit 0
	stdout, _, code := runVerify(t, nil, "go", "run", goDir, filepath.Join(r, "hooks/install.sh"))
	if code != 0 {
		t.Errorf("compress CLI skip path: want exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "Skipping") {
		t.Errorf("compress CLI skip path: expected 'Skipping' in output, got:\n%s", stdout)
	}

	// Error path: missing file → exit 1
	_, _, code = runVerify(t, nil, "go", "run", goDir, filepath.Join(r, "does-not-exist.md"))
	if code != 1 {
		t.Errorf("compress CLI missing-file path: want exit 1, got %d", code)
	}
}

// TestHookInstallFlow runs the full install → activate → mode-track → statusline → uninstall
// lifecycle in an isolated temp HOME directory.
func TestHookInstallFlow(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not on PATH")
	}
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not on PATH")
	}

	r := root()
	home := t.TempDir()
	claudeDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	existingSettings := map[string]any{
		"statusLine": map[string]any{"type": "command", "command": "bash /tmp/existing-statusline.sh"},
		"hooks": map[string]any{
			"Notification": []any{map[string]any{"hooks": []any{map[string]any{"type": "command", "command": "echo keep-me"}}}},
		},
	}
	writeJSON(t, filepath.Join(claudeDir, "settings.json"), existingSettings)

	// Fresh install
	if _, _, code := runVerify(t, map[string]string{"HOME": home}, "bash", "hooks/install.sh"); code != 0 {
		t.Fatal("install.sh failed")
	}
	settings := readJSON(t, filepath.Join(claudeDir, "settings.json"))
	hooks := settings["hooks"].(map[string]any)
	sl := settings["statusLine"].(map[string]any)
	if sl["command"] != "bash /tmp/existing-statusline.sh" {
		t.Error("install.sh clobbered existing statusLine")
	}
	if _, ok := hooks["SessionStart"]; !ok {
		t.Error("SessionStart hook missing after install")
	}
	if _, ok := hooks["UserPromptSubmit"]; !ok {
		t.Error("UserPromptSubmit hook missing after install")
	}

	// Activate — default mode
	stdout, _, _ := runVerify(t, map[string]string{"HOME": home}, "node", "hooks/caveman-activate.js")
	if !strings.Contains(stdout, "CAVEMAN MODE ACTIVE.") {
		t.Error("activation output missing caveman banner")
	}
	if strings.Contains(stdout, "STATUSLINE SETUP NEEDED") {
		t.Error("activation should stay quiet when custom statusline exists")
	}
	assertFlagFile(t, claudeDir, "full")

	// Activate — custom default mode via env var
	runVerify(t, map[string]string{"HOME": home, "CAVEMAN_DEFAULT_MODE": "ultra"}, "node", "hooks/caveman-activate.js")
	assertFlagFile(t, claudeDir, "ultra")

	// Activate — "off" mode removes flag
	runVerify(t, map[string]string{"HOME": home, "CAVEMAN_DEFAULT_MODE": "off"}, "node", "hooks/caveman-activate.js")
	if _, err := os.Stat(filepath.Join(claudeDir, ".caveman-active")); !os.IsNotExist(err) {
		t.Error("off mode should remove flag file")
	}

	// Mode tracker — /caveman with off default should not write flag
	runModeTracker(t, r, home, map[string]string{"CAVEMAN_DEFAULT_MODE": "off"}, `{"prompt":"/caveman"}`)
	if _, err := os.Stat(filepath.Join(claudeDir, ".caveman-active")); !os.IsNotExist(err) {
		t.Error("/caveman with off default should not write flag")
	}

	// Reset flag to full for remaining tests
	_ = os.WriteFile(filepath.Join(claudeDir, ".caveman-active"), []byte("full"), 0600)

	// Mode tracker — empty prompt (no-op)
	runModeTracker(t, r, home, nil, "")

	// Mode tracker — /caveman ultra
	trackerOut := runModeTracker(t, r, home, nil, `{"prompt":"/caveman ultra"}`)
	if trackerOut != "" {
		t.Error("mode tracker should stay silent on output")
	}
	assertFlagFile(t, claudeDir, "ultra")

	// Mode tracker — "normal mode" deactivates
	runModeTracker(t, r, home, nil, `{"prompt":"normal mode"}`)
	if _, err := os.Stat(filepath.Join(claudeDir, ".caveman-active")); !os.IsNotExist(err) {
		t.Error("'normal mode' should remove flag file")
	}

	// Statusline badge
	_ = os.WriteFile(filepath.Join(claudeDir, ".caveman-active"), []byte("wenyan-ultra"), 0600)
	stdout, _, _ = runVerify(t, map[string]string{"HOME": home}, "bash", "hooks/caveman-statusline.sh")
	if !strings.Contains(stdout, "[CAVEMAN:WENYAN-ULTRA]") {
		t.Errorf("statusline badge mismatch, got: %q", stdout)
	}

	// Idempotent reinstall
	stdout, _, _ = runVerify(t, map[string]string{"HOME": home}, "bash", "hooks/install.sh")
	if !strings.Contains(stdout, "Nothing to do") {
		t.Error("install.sh should be idempotent")
	}

	// Uninstall restores exactly original settings
	runVerify(t, map[string]string{"HOME": home}, "bash", "hooks/uninstall.sh")
	settingsAfter := readJSON(t, filepath.Join(claudeDir, "settings.json"))
	if fmt.Sprintf("%v", settingsAfter) != fmt.Sprintf("%v", existingSettings) {
		t.Errorf("uninstall.sh did not restore non-caveman settings\ngot:  %v\nwant: %v", settingsAfter, existingSettings)
	}
	if _, err := os.Stat(filepath.Join(claudeDir, ".caveman-active")); !os.IsNotExist(err) {
		t.Error("uninstall.sh should remove flag file")
	}

	// Fresh install + activate (no prior settings)
	home2 := t.TempDir()
	if _, _, code := runVerify(t, map[string]string{"HOME": home2}, "bash", "hooks/install.sh"); code != 0 {
		t.Fatal("fresh install.sh failed")
	}
	claudeDir2 := filepath.Join(home2, ".claude")
	s2 := readJSON(t, filepath.Join(claudeDir2, "settings.json"))
	if _, ok := s2["statusLine"]; !ok {
		t.Error("fresh install should configure statusLine")
	}
	stdout, _, _ = runVerify(t, map[string]string{"HOME": home2}, "node", "hooks/caveman-activate.js")
	if strings.Contains(stdout, "STATUSLINE SETUP NEEDED") {
		t.Error("fresh install should not nudge for statusline")
	}
	runVerify(t, map[string]string{"HOME": home2}, "bash", "hooks/uninstall.sh")
	s2after := readJSON(t, filepath.Join(claudeDir2, "settings.json"))
	if len(s2after) != 0 {
		t.Errorf("fresh uninstall should leave empty settings, got: %v", s2after)
	}
}

// ---------- helpers ----------

func assertFlagFile(t *testing.T, claudeDir, want string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(claudeDir, ".caveman-active"))
	if err != nil {
		t.Fatalf("flag file not found: %v", err)
	}
	if string(data) != want {
		t.Errorf("flag file: want %q, got %q", want, string(data))
	}
}

func runModeTracker(t *testing.T, repoRoot, home string, extra map[string]string, stdin string) string {
	t.Helper()
	cmd := exec.Command("node", "hooks/caveman-mode-tracker.js")
	cmd.Dir = repoRoot
	env := os.Environ()
	env = append(env, "HOME="+home, "USERPROFILE="+home)
	for k, v := range extra {
		env = append(env, k+"="+v)
	}
	cmd.Env = env
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Logf("mode tracker stderr: %s", ee.Stderr)
		}
	}
	return strings.TrimSpace(string(out))
}

// validateCompressed performs lightweight structural validation: heading count,
// code block count, and URL set must be preserved.
func validateCompressed(orig, comp string) []string {
	var errs []string

	origH := countHeadings(orig)
	compH := countHeadings(comp)
	if origH != compH {
		errs = append(errs, fmt.Sprintf("heading count mismatch: %d vs %d", origH, compH))
	}

	origBlocks := extractCodeBlocks(orig)
	compBlocks := extractCodeBlocks(comp)
	if len(origBlocks) != len(compBlocks) {
		errs = append(errs, fmt.Sprintf("code block count mismatch: %d vs %d", len(origBlocks), len(compBlocks)))
	} else {
		for i := range origBlocks {
			if origBlocks[i] != compBlocks[i] {
				errs = append(errs, fmt.Sprintf("code block %d content changed", i+1))
			}
		}
	}

	origURLs := extractURLs(orig)
	compURLs := extractURLs(comp)
	for u := range origURLs {
		if !compURLs[u] {
			errs = append(errs, fmt.Sprintf("URL lost: %s", u))
		}
	}
	return errs
}

var headingRe = regexp.MustCompile(`(?m)^#{1,6}\s+`)
var urlRe = regexp.MustCompile(`https?://[^\s)]+`)
var fenceRe = regexp.MustCompile("(?m)^(\\s{0,3})(`{3,}|~{3,})")

func countHeadings(text string) int {
	return len(headingRe.FindAllString(text, -1))
}

func extractURLs(text string) map[string]bool {
	m := make(map[string]bool)
	for _, u := range urlRe.FindAllString(text, -1) {
		m[u] = true
	}
	return m
}

func extractCodeBlocks(text string) []string {
	var blocks []string
	lines := strings.Split(text, "\n")
	i, n := 0, len(lines)
	for i < n {
		m := fenceRe.FindStringSubmatch(lines[i])
		if m == nil {
			i++
			continue
		}
		fenceChar := rune(m[2][0])
		fenceLen := len(m[2])
		blockLines := []string{lines[i]}
		i++
		closed := false
		for i < n {
			cm := fenceRe.FindStringSubmatch(lines[i])
			if cm != nil && rune(cm[2][0]) == fenceChar && len(cm[2]) >= fenceLen {
				blockLines = append(blockLines, lines[i])
				closed = true
				i++
				break
			}
			blockLines = append(blockLines, lines[i])
			i++
		}
		if closed {
			blocks = append(blocks, strings.Join(blockLines, "\n"))
		}
	}
	return blocks
}

func isNaturalLanguage(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".txt" || ext == ".rst" || ext == ".markdown"
}
