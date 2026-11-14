package godebugbar

import (
	"context"
	"log/slog"

	"github.com/DarkoKlisuric/godebugbar/internal/contextstate"
)

type SlogHandler struct {
	next   slog.Handler
	attrs  []slog.Attr
	groups []string
}

func NewSlogHandler(next slog.Handler) *SlogHandler {
	return &SlogHandler{next: next}
}

func (h *SlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if contextstate.RecorderFromContext(ctx) != nil {
		return true
	}

	return h.next != nil && h.next.Enabled(ctx, level)
}

func (h *SlogHandler) Handle(ctx context.Context, record slog.Record) error {
	if rec := contextstate.RecorderFromContext(ctx); rec != nil {
		rec.AddLog(h.entry(record))
	}
	if h.next != nil && h.next.Enabled(ctx, record.Level) {
		return h.next.Handle(ctx, record)
	}

	return nil
}

func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := h.clone()
	out.attrs = append(out.attrs, h.qualify(attrs)...)
	if h.next != nil {
		out.next = h.next.WithAttrs(attrs)
	}

	return out
}

func (h *SlogHandler) WithGroup(name string) slog.Handler {
	out := h.clone()
	out.groups = append(out.groups, name)
	if h.next != nil {
		out.next = h.next.WithGroup(name)
	}

	return out
}

func (h *SlogHandler) clone() *SlogHandler {
	return &SlogHandler{
		next:   h.next,
		attrs:  append([]slog.Attr(nil), h.attrs...),
		groups: append([]string(nil), h.groups...),
	}
}

func (h *SlogHandler) qualify(attrs []slog.Attr) []slog.Attr {
	if len(h.groups) == 0 {
		return attrs
	}
	out := make([]slog.Attr, len(attrs))
	for i, attr := range attrs {
		key := attr.Key
		for j := len(h.groups) - 1; j >= 0; j-- {
			key = h.groups[j] + "." + key
		}
		out[i] = slog.Attr{Key: key, Value: attr.Value}
	}

	return out
}

func (h *SlogHandler) entry(record slog.Record) contextstate.LogEntry {
	values := make(map[string]string, len(h.attrs)+record.NumAttrs())
	for _, attr := range h.attrs {
		values[attr.Key] = attr.Value.String()
	}
	record.Attrs(func(attr slog.Attr) bool {
		for _, q := range h.qualify([]slog.Attr{attr}) {
			values[q.Key] = q.Value.String()
		}
		return true
	})

	return contextstate.LogEntry{
		Time:    record.Time,
		Level:   record.Level.String(),
		Message: record.Message,
		Attrs:   values,
	}
}
