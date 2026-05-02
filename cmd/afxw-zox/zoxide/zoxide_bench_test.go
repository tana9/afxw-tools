package zoxide

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeBenchInput(b *testing.B, n int) string {
	b.Helper()
	tmpDir := b.TempDir()
	var sb strings.Builder
	for i := range n {
		dir := filepath.Join(tmpDir, fmt.Sprintf("dir%03d", i))
		if err := os.MkdirAll(dir, 0755); err != nil {
			b.Fatal(err)
		}
		fmt.Fprintf(&sb, "%.1f %s\n", float64(n-i)*0.5, dir)
	}
	return sb.String()
}

func BenchmarkParseQueryOutput_50(b *testing.B) {
	input := makeBenchInput(b, 50)
	b.ResetTimer()
	for b.Loop() {
		if _, err := parseQueryOutput(input); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseQueryOutput_200(b *testing.B) {
	input := makeBenchInput(b, 200)
	b.ResetTimer()
	for b.Loop() {
		if _, err := parseQueryOutput(input); err != nil {
			b.Fatal(err)
		}
	}
}
