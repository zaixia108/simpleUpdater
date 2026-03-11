package tools

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
)

var levelColors = map[slog.Level]string{
	slog.LevelDebug: colorBlue,
	slog.LevelInfo:  colorGreen,
	slog.LevelWarn:  colorYellow,
	slog.LevelError: colorRed,
}

var levelStrings = map[slog.Level]string{
	slog.LevelDebug: "DEBUG",
	slog.LevelInfo:  "INFO ",
	slog.LevelWarn:  "WARN ",
	slog.LevelError: "ERROR",
}

type ColorHandler struct {
	w       io.Writer
	level   *slog.LevelVar
	attrs   []slog.Attr
	groups  []string
	mu      sync.Mutex
	noColor bool
}

type LoggerStruct struct {
	*slog.Logger
	levelVar *slog.LevelVar
}

type LogConfig struct {
	Level      string
	Output     string
	FilePath   string
	MaxSize    int
	MaxBackups int
}

var Logger *LoggerStruct
var logLevel *slog.LevelVar

func init() {
	logLevel = &slog.LevelVar{}
	logLevel.Set(slog.LevelInfo)

	handler := NewColorHandler(os.Stdout, logLevel, false)
	Logger = &LoggerStruct{
		Logger:   slog.New(handler),
		levelVar: logLevel,
	}
}

func NewColorHandler(w io.Writer, level *slog.LevelVar, noColor bool) *ColorHandler {
	return &ColorHandler{
		w:       w,
		level:   level,
		noColor: noColor,
	}
}

func (h *ColorHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *ColorHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var buf []byte

	timeStr := r.Time.Format("2006-01-02 15:04:05.000")
	levelStr := levelStrings[r.Level]
	color := levelColors[r.Level]

	if !h.noColor {
		buf = append(buf, colorGray...)
	}
	buf = append(buf, timeStr...)
	buf = append(buf, " "...)

	if !h.noColor {
		buf = append(buf, color...)
	}
	buf = append(buf, levelStr...)
	buf = append(buf, " "...)

	if !h.noColor {
		buf = append(buf, colorReset...)
	}

	if r.PC != 0 {
		frame, _ := runtime.CallersFrames([]uintptr{r.PC}).Next()
		if frame.File != "" {
			file := filepath.Base(frame.File)
			if !h.noColor {
				buf = append(buf, colorCyan...)
			}
			buf = append(buf, fmt.Sprintf("[%s:%d] ", file, frame.Line)...)
			if !h.noColor {
				buf = append(buf, colorReset...)
			}
		}
	}

	buf = append(buf, r.Message...)

	r.Attrs(func(a slog.Attr) bool {
		buf = append(buf, " "...)
		buf = h.appendAttr(buf, a)
		return true
	})

	for _, attr := range h.attrs {
		buf = append(buf, " "...)
		buf = h.appendAttr(buf, attr)
	}

	buf = append(buf, '\n')

	_, err := h.w.Write(buf)
	return err
}

func (h *ColorHandler) appendAttr(buf []byte, a slog.Attr) []byte {
	if a.Equal(slog.Attr{}) {
		return buf
	}

	if !h.noColor {
		buf = append(buf, colorBlue...)
	}
	buf = append(buf, a.Key...)
	if !h.noColor {
		buf = append(buf, colorReset...)
	}
	buf = append(buf, '=')

	switch a.Value.Kind() {
	case slog.KindString:
		buf = append(buf, fmt.Sprintf("%q", a.Value.String())...)
	case slog.KindAny:
		buf = append(buf, fmt.Sprintf("%+v", a.Value.Any())...)
	default:
		buf = append(buf, a.Value.String()...)
	}

	return buf
}

func (h *ColorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	h2 := *h
	h2.attrs = append(h2.attrs, attrs...)
	return &h2
}

func (h *ColorHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := *h
	h2.groups = append(h2.groups, name)
	return &h2
}

func InitWithConfig(cfg LogConfig) error {
	logLevel = &slog.LevelVar{}

	switch cfg.Level {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "info":
		logLevel.Set(slog.LevelInfo)
	case "warn":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	default:
		logLevel.Set(slog.LevelInfo)
	}

	var writers []io.Writer

	switch cfg.Output {
	case "console":
		writers = append(writers, os.Stdout)
	case "file":
		fileWriter, err := getFileWriter(cfg)
		if err != nil {
			return err
		}
		writers = append(writers, fileWriter)
	case "both":
		writers = append(writers, os.Stdout)
		fileWriter, err := getFileWriter(cfg)
		if err != nil {
			return err
		}
		writers = append(writers, fileWriter)
	default:
		writers = append(writers, os.Stdout)
	}

	var handlers []slog.Handler
	for i, w := range writers {
		noColor := cfg.Output == "file" && i > 0
		handlers = append(handlers, NewColorHandler(w, logLevel, noColor))
	}

	multiHandler := NewMultiHandler(handlers...)
	Logger = &LoggerStruct{
		Logger:   slog.New(multiHandler),
		levelVar: logLevel,
	}

	return nil
}

func getFileWriter(cfg LogConfig) (io.Writer, error) {
	if cfg.FilePath == "" {
		cfg.FilePath = "logs/app.log"
	}

	dir := filepath.Dir(cfg.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func SetLevel(level string) {
	switch level {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "info":
		logLevel.Set(slog.LevelInfo)
	case "warn":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	}
}

type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, r); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return NewMultiHandler(newHandlers...)
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return NewMultiHandler(newHandlers...)
}
