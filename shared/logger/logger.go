package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"shared/config"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	LevelDebug = 0
	LevelInfo  = 1
	LevelWarn  = 2
	LevelError = 3
)

var currentLevel = LevelInfo

func Init(cfg config.LogConfig) error {
	currentLevel = parseLevel(cfg.Level)

	out := io.Writer(os.Stdout)
	var fileErr error
	if cfg.File != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.File), 0o755); err != nil {
			fileErr = err
		} else {
			fileWriter := &lumberjack.Logger{
				Filename:   cfg.File,
				MaxSize:    cfg.MaxSize,
				MaxBackups: cfg.MaxBackups,
				MaxAge:     cfg.MaxAge,
				Compress:   cfg.Compress,
			}
			out = io.MultiWriter(os.Stdout, fileWriter)
		}
	}

	log.SetOutput(out)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("")

	if fileErr != nil {
		Errorf("日志文件不可用: %v", fileErr)
		return nil
	}
	Infof("日志已启用 file=%s level=%s", cfg.File, cfg.Level)
	return nil
}

func Debugf(format string, args ...any) { print(LevelDebug, "DEBUG", format, args...) }
func Infof(format string, args ...any)  { print(LevelInfo, "INFO", format, args...) }
func Warnf(format string, args ...any)  { print(LevelWarn, "WARN", format, args...) }
func Errorf(format string, args ...any) { print(LevelError, "ERROR", format, args...) }

func Fatalf(format string, args ...any) {
	print(LevelError, "FATAL", format, args...)
	os.Exit(1)
}

func print(level int, name, format string, args ...any) {
	if level < currentLevel {
		return
	}
	msg := fmt.Sprintf(format, args...)
	log.Output(3, "["+name+"] "+msg)
}

func parseLevel(s string) int {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return LevelDebug
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}
