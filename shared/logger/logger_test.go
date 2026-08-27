package logger

import (
	"os"
	"path/filepath"
	"testing"

	"shared/config"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{" info ", LevelInfo},
		{"warn", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"", LevelInfo},
		{"nope", LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := parseLevel(tt.in)
			if got != tt.want {
				t.Fatalf("parseLevel(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestInit_stdoutOnly(t *testing.T) {
	if err := Init(config.LogConfig{Level: "error"}); err != nil {
		t.Fatal(err)
	}
	if currentLevel != LevelError {
		t.Fatalf("level=%d", currentLevel)
	}
	Debug("should be dropped")
	Info("should be dropped")
	Warn("should be dropped")
	Error("visible")
}

func TestInit_withFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "t.log")
	if err := Init(config.LogConfig{
		Level:      "info",
		File:       path,
		MaxSize:    1,
		MaxBackups: 1,
		MaxAge:     1,
	}); err != nil {
		t.Fatal(err)
	}
	Info("hello file")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected log file to be written")
	}
}

func TestInit_mkdirFail(t *testing.T) {
	dir := t.TempDir()
	block := filepath.Join(dir, "notdir")
	if err := os.WriteFile(block, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := Init(config.LogConfig{
		Level: "info",
		File:  filepath.Join(block, "nested", "a.log"),
	})
	if err != nil {
		t.Fatal(err)
	}
}
