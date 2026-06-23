package utils

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

func SetupLogger() {
	logger := slog.New(NewDaLog(
		os.Stdout,
		DaLogStyleLongType1,
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	))
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelDebug)
}
func init() {
	SetupLogger()
}

func NewDaLog(w io.Writer, s func(rec slog.Record) string, opts *slog.HandlerOptions) slog.Handler {
	return &DaLog{
		defaultHandler: slog.NewTextHandler(w, opts),
		Writer:         w,
		HandlerOptions: opts,
		style:          s,
		groups:         []string{},
	}
}

type DaLog struct {
	*slog.HandlerOptions
	defaultHandler slog.Handler
	Writer         io.Writer
	style          func(rec slog.Record) string
	groups         []string
}

// Enabled implements slog.Handler.
func (d *DaLog) Enabled(ctx context.Context, lvl slog.Level) bool {
	return lvl >= d.Level.Level()
}

// Handle implements slog.Handler.
func (d *DaLog) Handle(ctx context.Context, rec slog.Record) error {
	toPrint := d.style(rec)
	_, err := fmt.Fprintln(d.Writer, toPrint)
	if err != nil {
		return err
	}
	return nil
}

// WithAttrs implements slog.Handler.
func (d *DaLog) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &DaLog{
		defaultHandler: d.defaultHandler.WithAttrs(attrs),
		Writer:         d.Writer,
		HandlerOptions: d.HandlerOptions,
		style:          d.style,
	}
}

// WithGroup implements slog.Handler.
func (d *DaLog) WithGroup(name string) slog.Handler {
	newGroups := append(d.groups, name)
	return &DaLog{
		defaultHandler: d.defaultHandler.WithGroup(name),
		Writer:         d.Writer,
		HandlerOptions: d.HandlerOptions,
		style:          d.style,
		groups:         newGroups,
	}
}

func ColorByLevel(text string, rec slog.Level, isBold, isBG bool) string {
	if isBold {
		text = MakeBold(text)
	}
	if isBG {
		switch rec {
		case slog.LevelDebug:
			return ChangeColor(text, 46)
		case slog.LevelInfo:
			return ChangeColor(text, 42)
		case slog.LevelWarn:
			return ChangeColor(text, 43)
		case slog.LevelError:
			return ChangeColor(text, 41)
		default:
			return rec.String()
		}
	}
	switch rec {
	case slog.LevelDebug:
		return ChangeColor(text, 36)
	case slog.LevelInfo:
		return ChangeColor(text, 32)
	case slog.LevelWarn:
		return ChangeColor(text, 33)
	case slog.LevelError:
		return ChangeColor(text, 31)
	default:
		return rec.String()
	}
}

func ChangeColor(str string, color int) string {
	return fmt.Sprintf("\033[%vm%v\033[0m", color, str)
}

func MakeBold(str string) string {
	return fmt.Sprintf("\033[1m%v\033[0m", str)
}

func DaLogStyleLongType1(rec slog.Record) string {
	// Extract the source file and line number
	wd, _ := os.Getwd()
	source := ""
	if rec.PC != 0 {
		fn := runtime.FuncForPC(rec.PC)
		if fn != nil {
			file, line := fn.FileLine(rec.PC)
			source = fmt.Sprintf("%s:%d", strings.TrimPrefix(file, wd+"/"), line)
		}
	}

	// Build the log entry
	result := fmt.Sprintf(
		"[%v]\t%v\t%v\t%v\t",
		rec.Time.Format("2006-01-02 15:04:05"),
		ColorByLevel(rec.Level.String(), rec.Level, true, false),
		source, // Include source info
		rec.Message,
	)
	rec.Attrs(func(a slog.Attr) bool {
		result += fmt.Sprintf("%v: %v\t", ColorByLevel(a.Key, rec.Level, false, false), a.Value)
		return true
	})
	return result
}
