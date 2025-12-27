package logger

import (
	"context"
	"log/slog"
)

type ctxKey struct{}

var attrsKey = ctxKey{} // более понятное имя переменной

// AddToCtx adds attributes to the context.
func AddToCtx(ctx context.Context, attrs ...slog.Attr) context.Context {
	if len(attrs) == 0 { // проверка на пустой список
		return ctx
	}

	existing, _ := ctx.Value(attrsKey).([]slog.Attr)
	// можно убрать проверку ok, т.к. при false вернётся nil slice

	// создаём новый слайс чтобы избежать race conditions
	combined := make([]slog.Attr, 0, len(existing)+len(attrs))
	combined = append(combined, existing...)
	combined = append(combined, attrs...)

	return context.WithValue(ctx, attrsKey, combined)
}

// GetFromCtx извлекает атрибуты из контекста (опционально, для тестов)
func GetFromCtx(ctx context.Context) []slog.Attr {
	attrs, _ := ctx.Value(attrsKey).([]slog.Attr)
	return attrs
}

// ContextHandler is a slog.Handler that extracts attributes from context.
type ContextHandler struct {
	slog.Handler
}

// Handle extracts attributes from context and adds them to the record.
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(attrsKey).([]slog.Attr); ok && len(attrs) > 0 {
		r.AddAttrs(attrs...)
	}

	return h.Handler.Handle(ctx, r)
}

// NewContextHandler создаёт новый handler с поддержкой контекста
func NewContextHandler(h slog.Handler) *ContextHandler {
	return &ContextHandler{Handler: h}
}
