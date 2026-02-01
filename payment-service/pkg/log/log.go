// Package log is a package that implements handler for logger
package log

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"runtime"
	"time"
)

// better local handler is a handler for slog
// for colorful and more clearly output
type betterLocalHandler struct {
	slog.Handler
	log *log.Logger
}

func (b *betterLocalHandler) Handle(ctx context.Context, r slog.Record) error {
	lvl := makeLvlColorful(r.Level)
	timestamp := r.Time.Format("15:04:05")

	if r.PC != 0 && r.Level != slog.LevelInfo {
		frames := runtime.CallersFrames([]uintptr{r.PC})
		frame, _ := frames.Next()
		r.AddAttrs(
			slog.String("file", frame.File),
			slog.String("function", frame.Function),
			slog.Int("line", frame.Line),
		)
	}

	attrs := getJSONAttrs(r)

	msg := fmt.Sprintf("[%s] %s - %s",
		timestamp, lvl,
		r.Message)

	if len(attrs) != 0 {
		msg += " " + string(attrs)
	}

	b.log.Println(msg)

	return nil
}

func SetupHandler(out io.Writer, env string) (slog.Handler, error) {
	switch env {
	case "prod":
		return slog.NewJSONHandler(out, &slog.HandlerOptions{
			AddSource: true,
		}), nil
	case "dev", "local":
		return &betterLocalHandler{
			Handler: slog.NewTextHandler(out, &slog.HandlerOptions{
				AddSource: true,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey {
						return slog.String("time", time.Now().Format("15:04:05"))
					}
					return a
				},
				Level: slog.LevelDebug,
			}),
			log: log.New(out, "", 0),
		}, nil
	}

	return nil, errors.New("invalid type of env")
}

func makeLvlColorful(lvl slog.Level) string {
	switch lvl {
	case slog.LevelDebug:
		return fmt.Sprintf("\033[37m%s\033[0m", lvl.String())
	case slog.LevelWarn:
		return fmt.Sprintf("\033[93m%s\033[0m", lvl.String())
	case slog.LevelError:
		return fmt.Sprintf("\033[31m%s\033[0m", lvl.String())
	case slog.LevelInfo:
		return fmt.Sprintf("\033[32m%s\033[0m", lvl.String())
	}
	return lvl.String()
}

func getJSONAttrs(r slog.Record) []byte {
	fields := make(map[string]any, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	attrs := make([]byte, 0)
	var err error

	if len(fields) != 0 {
		attrs, err = json.MarshalIndent(fields, "", " ")
		if err != nil {
			return nil
		}
	}

	return attrs
}
