package audit

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var forbiddenExcelizeCalls = []string{
	".Save(",
	".Write(",
	".SetCell",
	".SetSheet",
	".Update",
}

func TestAnalyzerDoesNotUseExcelizeMutationAPIs(t *testing.T) {
	root := analyzerSourceRoot(t)
	var violations []string

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(content)
		for _, needle := range forbiddenExcelizeCalls {
			if strings.Contains(text, needle) {
				violations = append(violations, path+": contains "+needle)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk analyzer sources: %v", err)
	}
	if len(violations) > 0 {
		t.Fatalf("mutation API guard failed:\n%s", strings.Join(violations, "\n"))
	}
}

func analyzerSourceRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename)))
}
