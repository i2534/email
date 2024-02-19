package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

type SimpleHandler struct {
	lv slog.Level
	mu *sync.Mutex
	w  io.Writer
}

func NewSimpleHandler(w io.Writer, level slog.Level) *SimpleHandler {
	return &SimpleHandler{
		lv: level,
		mu: &sync.Mutex{},
		w:  w,
	}
}

func (h *SimpleHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.lv
}

func (h *SimpleHandler) Handle(_ context.Context, r slog.Record) error {
	buf := NewBuffer()
	defer buf.Free()
	// time
	if !r.Time.IsZero() {
		*buf = r.Time.AppendFormat(*buf, "2006-01-02 15:04:05")
		buf.WriteString(" ")
	}

	// level
	buf.WriteString(fmt.Sprintf("% 5s", r.Level.String()))

	// message
	buf.WriteString(" ")
	buf.WriteString(r.Message)

	// attrs
	if r.NumAttrs() > 0 {
		r.Attrs(func(a slog.Attr) bool {
			buf.WriteByte(' ')
			buf.WriteString(a.Key)
			buf.WriteByte('=')
			buf.WriteString(fmt.Sprintf("%q", a.Value))
			return true
		})
	}

	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()

	_, e := h.w.Write(*buf)
	return e
}

func (h *SimpleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	//TODO: implement
	return h
}

func (h *SimpleHandler) WithGroup(name string) slog.Handler {
	//TODO: implement
	return h
}
