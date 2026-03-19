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
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

var (
	cacheDirOnce sync.Once
	cacheDirPath string
)

func goCacheDir(tb testing.TB) string {
	tb.Helper()

	cacheDirOnce.Do(func() {
		cacheDirPath = filepath.Join(repoRoot(tb), ".gocache", "mock-kling")
		if err := os.MkdirAll(cacheDirPath, 0o755); err != nil {
			tb.Fatalf("failed to create mock go cache: %v", err)
		}
	})
	return cacheDirPath
}

func RunExample(tb testing.TB, examplePath string, args ...string) string {
	tb.Helper()

	cmdArgs := append([]string{"run", examplePath}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = repoRoot(tb)
	cmd.Env = append(os.Environ(),
		"GOCACHE="+goCacheDir(tb),
		"QINIU_API_KEY=test-key",
		"QINIU_MOCK_CURL=1",
	)

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
