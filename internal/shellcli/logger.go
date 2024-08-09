package shellcli

import (
	"context"
	"encoding/json"
	"github.com/fatih/color"
	"io"
	"log"
	"log/slog"
)

type PrettyHandlerOptions struct {
	// wrap slog handler options
	SlogOpts slog.HandlerOptions
}

type PrettyHandler struct {
	// wrap slog handler
	slog.Handler
	// wrap logger from log
	logger *log.Logger
}

// Handle
// Implementation for PrettyHandler type
// in the following implementation, we define the handler to set a color for each different level
// Here, the logger slog record is used : the record is the format of each field you want to be
// logged by the handler. It can contain the timestamp, the source file, the log level, etc ...
func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	// Set the log level in the record
	level := r.Level.String() + ":"
	// set the color of the level printed in the console
	switch r.Level {
	case slog.LevelInfo:
		level = color.GreenString(level)
	case slog.LevelDebug:
		level = color.CyanString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}
	// Set the timestamp formatter
	timeStr := r.Time.Format("[15:04:05.000]")
	// Set the message to be printed in white
	msg := color.WhiteString(r.Message)

	// Set any additional field the developer may add
	fields := make(map[string]interface{}, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})
	b, err := json.Marshal(fields)
	if err != nil {
		// error if the values cannot be parsed in json
		return err
	}

	// Set the record specifying the order of each part of the record
	h.logger.Println(timeStr, level, msg, string(b))
	return nil
}

// NewPrettyHandler
// create a new PrettyHandler
func NewPrettyHandler(out io.Writer, opts PrettyHandlerOptions) *PrettyHandler {
	return &PrettyHandler{
		Handler: slog.NewTextHandler(out, &opts.SlogOpts),
		logger:  log.New(out, "", 0),
	}
}
