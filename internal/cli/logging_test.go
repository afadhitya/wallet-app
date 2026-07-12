package cli

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestMultiHandlerHandle(t *testing.T) {
	buf1 := new(bytes.Buffer)
	buf2 := new(bytes.Buffer)
	h1 := slog.NewTextHandler(buf1, nil)
	h2 := slog.NewTextHandler(buf2, nil)
	mh := NewMultiHandler(h1, h2)

	logger := slog.New(mh)
	logger.Info("hello")

	output1 := buf1.String()
	output2 := buf2.String()

	if output1 == "" {
		t.Error("expected output in first handler")
	}
	if output2 == "" {
		t.Error("expected output in second handler")
	}
}

func TestMultiHandlerEnabled(t *testing.T) {
	buf := new(bytes.Buffer)
	h1 := slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelError})
	h2 := slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	mh := NewMultiHandler(h1, h2)

	if !mh.Enabled(context.Background(), slog.LevelError) {
		t.Error("expected Enabled=true for error level (h1 accepts error)")
	}
	if !mh.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("expected Enabled=true for info level (h2 accepts info)")
	}
	if mh.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("expected Enabled=false for debug level (neither accepts debug)")
	}
}

func TestMultiHandlerWithAttrs(t *testing.T) {
	buf := new(bytes.Buffer)
	h := slog.NewTextHandler(buf, nil)
	mh := NewMultiHandler(h)

	mh2 := mh.WithAttrs([]slog.Attr{slog.String("key", "value")})
	if mh2 == nil {
		t.Error("expected non-nil handler from WithAttrs")
	}

	logger := slog.New(mh2)
	logger.Info("test")
	if !bytes.Contains(buf.Bytes(), []byte("key=value")) {
		t.Error("expected 'key=value' in log output")
	}
}

func TestMultiHandlerWithGroup(t *testing.T) {
	buf := new(bytes.Buffer)
	h := slog.NewTextHandler(buf, nil)
	mh := NewMultiHandler(h)

	mh2 := mh.WithGroup("group")
	if mh2 == nil {
		t.Error("expected non-nil handler from WithGroup")
	}

	logger := slog.New(mh2)
	logger.Info("test", slog.String("key", "value"))
	if !bytes.Contains(buf.Bytes(), []byte("group")) {
		t.Error("expected 'group' in log output")
	}
}

func TestMultiHandlerHandleFail(t *testing.T) {
	buf := new(bytes.Buffer)
	h1 := slog.NewTextHandler(buf, nil)
	h2 := &failingHandler{}
	mh := NewMultiHandler(h1, h2)

	logger := slog.New(mh)
	logger.Info("hello")

	if buf.String() == "" {
		t.Error("expected first handler to still write even though second fails")
	}
}

type failingHandler struct{}

func (f *failingHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (f *failingHandler) Handle(_ context.Context, _ slog.Record) error {
	return &testError{msg: "mock handler failure"}
}
func (f *failingHandler) WithAttrs(_ []slog.Attr) slog.Handler  { return f }
func (f *failingHandler) WithGroup(_ string) slog.Handler        { return f }

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

func TestNewLoggerDefaultWarn(t *testing.T) {
	dir := t.TempDir()

	cmd := &cobra.Command{}
	logger := newLogger(cmd, dir)
	logger.Info("should not appear")
	logger.Warn("should appear")

	fileContents, err := os.ReadFile(filepath.Join(dir, "wallet.log"))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	if bytes.Contains(fileContents, []byte("should not appear")) {
		t.Error("INFO level message should not appear at default WARN level")
	}
	if !bytes.Contains(fileContents, []byte("should appear")) {
		t.Error("WARN level message should appear at default WARN level")
	}
}

func TestNewLoggerVerboseInfo(t *testing.T) {
	dir := t.TempDir()

	cmd := &cobra.Command{}
	cmd.Flags().CountP("verbose", "v", "")
	_ = cmd.Flags().Set("verbose", "1")
	logger := newLogger(cmd, dir)
	logger.Info("info message")
	logger.Debug("debug message")

	fileContents, err := os.ReadFile(filepath.Join(dir, "wallet.log"))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	if !bytes.Contains(fileContents, []byte("info message")) {
		t.Error("INFO message should appear at INFO level")
	}
	if bytes.Contains(fileContents, []byte("debug message")) {
		t.Error("DEBUG message should not appear at INFO level")
	}
}

func TestNewLoggerVerboseDebug(t *testing.T) {
	dir := t.TempDir()

	cmd := &cobra.Command{}
	cmd.Flags().CountP("verbose", "v", "")
	_ = cmd.Flags().Set("verbose", "2")
	logger := newLogger(cmd, dir)
	logger.Debug("debug message")

	fileContents, err := os.ReadFile(filepath.Join(dir, "wallet.log"))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	if !bytes.Contains(fileContents, []byte("debug message")) {
		t.Error("DEBUG message should appear at DEBUG level")
	}
}

func TestNewLoggerVerboseBeyondMax(t *testing.T) {
	dir := t.TempDir()

	cmd := &cobra.Command{}
	cmd.Flags().CountP("verbose", "v", "")
	_ = cmd.Flags().Set("verbose", "5")
	logger := newLogger(cmd, dir)
	logger.Debug("debug message")

	fileContents, err := os.ReadFile(filepath.Join(dir, "wallet.log"))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	if !bytes.Contains(fileContents, []byte("debug message")) {
		t.Error("DEBUG message should appear when verbosity exceeds max")
	}
}

func TestNewLoggerWithLogFile(t *testing.T) {
	dir := t.TempDir()
	customPath := filepath.Join(dir, "custom.log")

	cmd := &cobra.Command{}
	cmd.Flags().CountP("verbose", "v", "")
	cmd.Flags().String("log-file", "", "")
	_ = cmd.Flags().Set("log-file", customPath)
	logger := newLogger(cmd, dir)
	logger.Warn("test warning")

	fileContents, err := os.ReadFile(customPath)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !bytes.Contains(fileContents, []byte("test warning")) {
		t.Errorf("expected log file to contain message, got: %s", string(fileContents))
	}

	defaultPath := filepath.Join(dir, "wallet.log")
	if _, err := os.Stat(defaultPath); !os.IsNotExist(err) {
		t.Error("expected default log file not to be created when --log-file is set")
	}
}

func TestNewLoggerLogFileOpenError(t *testing.T) {
	dir := "/nonexistent"

	cmd := &cobra.Command{}
	cmd.Flags().CountP("verbose", "v", "")
	cmd.Flags().String("log-file", "", "")
	_ = cmd.Flags().Set("log-file", "/nonexistent/dir/log.json")
	logger := newLogger(cmd, dir)
	logger.Warn("fallback")
}

func TestNewLoggerVerboseFlagError(t *testing.T) {
	dir := t.TempDir()

	cmd := &cobra.Command{}
	logger := newLogger(cmd, dir)
	logger.Warn("should appear")

	fileContents, err := os.ReadFile(filepath.Join(dir, "wallet.log"))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	if !bytes.Contains(fileContents, []byte("should appear")) {
		t.Error("logger should work without verbose flag registered")
	}
}
