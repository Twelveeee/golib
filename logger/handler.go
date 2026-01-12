package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/Twelveeee/golib/pool"
)

// DefaultHandler 自定义日志格式的 Handler
type DefaultHandler struct {
	w     io.Writer
	level slog.Level
	attrs []slog.Attr
	group string
	mu    sync.Mutex
}

// NewDefaultHandler 创建自定义格式的 Handler
func NewDefaultHandler(w io.Writer, level slog.Level) *DefaultHandler {
	return &DefaultHandler{
		w:     w,
		level: level,
	}
}

func (h *DefaultHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *DefaultHandler) Handle(ctx context.Context, r slog.Record) error {
	buf := pool.GlobalBytesPool.Get()
	defer pool.GlobalBytesPool.Put(buf)

	// 添加日志级别
	buf.WriteString(r.Level.String())
	buf.WriteString(": ")

	t := r.Time.Format("2006-01-02 15:04:05")
	buf.WriteString(t)
	buf.WriteByte(' ')

	// 添加 caller 信息
	if r.PC != 0 {
		if writeCallerWithSkip(buf, 4) {
			buf.WriteByte(' ')
		}
	}

	// 从 context 中提取 traceID
	if ctx != nil {
		if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
			buf.WriteString("traceID=")
			buf.WriteString(traceID)
			buf.WriteByte(' ')
		}
	}

	// 添加消息
	if r.Message != "" {
		buf.WriteString("msg=")
		buf.WriteString(r.Message)
	}

	// 添加预设的属性
	for _, attr := range h.attrs {
		buf.WriteByte(' ')
		h.appendAttr(buf, attr)
	}

	// 添加记录中的属性
	r.Attrs(func(attr slog.Attr) bool {
		buf.WriteByte(' ')
		h.appendAttr(buf, attr)
		return true
	})

	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(buf.Bytes())
	return err
}

func (h *DefaultHandler) appendAttr(buf *bytes.Buffer, attr slog.Attr) {
	// 处理分组
	if h.group != "" {
		buf.WriteString(h.group)
		buf.WriteByte('.')
	}

	buf.WriteString(attr.Key)
	buf.WriteByte('=')

	// 根据值类型格式化
	switch attr.Value.Kind() {
	case slog.KindString:
		buf.WriteString(attr.Value.String())
	case slog.KindInt64:
		fmt.Fprintf(buf, "%d", attr.Value.Int64())
	case slog.KindUint64:
		fmt.Fprintf(buf, "%d", attr.Value.Uint64())
	case slog.KindFloat64:
		fmt.Fprintf(buf, "%g", attr.Value.Float64())
	case slog.KindBool:
		fmt.Fprintf(buf, "%t", attr.Value.Bool())
	case slog.KindDuration:
		fmt.Fprint(buf, attr.Value.Duration())
	case slog.KindTime:
		buf.WriteString(attr.Value.Time().Format(time.DateTime))
	default:
		fmt.Fprint(buf, attr.Value.Any())
	}
}

func (h *DefaultHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	newAttrs = append(newAttrs, h.attrs...)
	newAttrs = append(newAttrs, attrs...)

	return &DefaultHandler{
		w:     h.w,
		level: h.level,
		attrs: newAttrs,
		group: h.group,
	}
}

func (h *DefaultHandler) WithGroup(name string) slog.Handler {
	newGroup := name
	if h.group != "" {
		newGroup = h.group + "." + name
	}

	return &DefaultHandler{
		w:     h.w,
		level: h.level,
		attrs: h.attrs,
		group: newGroup,
	}
}
