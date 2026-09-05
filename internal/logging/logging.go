// Package logging builds the application's zap logger.
package logging

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// maxLogAgeDays is how long a rotated log file is kept before lumberjack
// deletes it.
const maxLogAgeDays = 30

// New builds a structured logger that writes human-readable log lines to
// both standard output and the text file at path (appending to it, creating
// the file and any missing parent directory if needed). The file is rotated
// and pruned by lumberjack, which deletes rotated files older than
// maxLogAgeDays.
func New(path string) (*zap.Logger, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("creating log directory %q: %w", dir, err)
		}
	}

	fileWriter := &lumberjack.Logger{
		Filename: path,
		MaxAge:   maxLogAgeDays,
	}

	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zap.InfoLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(fileWriter), zap.InfoLevel),
	)

	return zap.New(core), nil
}
