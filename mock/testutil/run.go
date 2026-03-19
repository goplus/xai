package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func repoRoot(tb testing.TB) string {
	tb.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		tb.Fatal("failed to resolve testutil caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

var (
	cacheDirOnce sync.Once
	cacheDirPath string
)

func goCacheDir(tb testing.TB) string {
	tb.Helper()

	cacheDirOnce.Do(func() {
		cacheDirPath = filepath.Join(repoRoot(tb), ".gocache", "mock")
		if err := os.MkdirAll(cacheDirPath, 0o755); err != nil {
			tb.Fatalf("failed to create mock go cache: %v", err)
		}
	})
	return cacheDirPath
}

func RunExample(tb testing.TB, examplePath string, args ...string) string {
	tb.Helper()
	return RunExampleWithEnv(tb, examplePath, nil, args...)
}

func RunExampleWithEnv(tb testing.TB, examplePath string, env map[string]string, args ...string) string {
	tb.Helper()

	cmdArgs := append([]string{"run", examplePath}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = repoRoot(tb)
	cmd.Env = mergeEnv(os.Environ(), map[string]string{
		"GOCACHE": goCacheDir(tb),
	}, env)

	out, err := cmd.CombinedOutput()
	if err != nil {
		tb.Fatalf("go %s failed: %v\n%s", strings.Join(cmdArgs, " "), err, out)
	}
	return string(out)
}

func RequireContainsAll(tb testing.TB, output string, wants ...string) {
	tb.Helper()
	for _, want := range wants {
		if !strings.Contains(output, want) {
			tb.Fatalf("mock output missing %q\n%s", want, output)
		}
	}
}

func mergeEnv(base []string, overlays ...map[string]string) []string {
	envMap := make(map[string]string, len(base))
	for _, item := range base {
		if key, val, ok := strings.Cut(item, "="); ok {
			envMap[key] = val
		}
	}
	for _, overlay := range overlays {
		for key, val := range overlay {
			envMap[key] = val
		}
	}
	out := make([]string, 0, len(envMap))
	for key, val := range envMap {
		out = append(out, key+"="+val)
	}
	return out
}
