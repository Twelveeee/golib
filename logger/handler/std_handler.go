package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/Twelveeee/golib/constant"
	"github.com/Twelveeee/golib/pool"
)

// ANSI 颜色代码
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
	colorCyan   = "\033[36m"
)

// StdHandler 带颜色输出的 Handler
type StdHandler struct {
	w     io.Writer
	level slog.Level
	attrs []slog.Attr
	group string
	mu    sync.Mutex
}

// NewStdHandler 创建带颜色的 Handler
func NewStdHandler(w io.Writer, level slog.Level) *StdHandler {
	return &StdHandler{
		w:     w,
		level: level,
	}
}

func (h *StdHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *StdHandler) Handle(ctx context.Context, r slog.Record) error {
	buf := pool.GlobalBytesPool.Get()
	defer pool.GlobalBytesPool.Put(buf)

	// 根据日志级别选择颜色
	levelColor := h.getLevelColor(r.Level)

	// 添加日志级别(带颜色)
	buf.WriteString(levelColor)
	buf.WriteString(r.Level.String())
	buf.WriteString(colorReset)
	buf.WriteString(": ")

	// 添加时间(灰色)
	buf.WriteString(colorGray)
	t := r.Time.Format("2006-01-02 15:04:05")
	buf.WriteString(t)
	buf.WriteString(colorReset)
	buf.WriteByte(' ')

	// 添加 caller 信息(青色)
	if r.PC != 0 {
		buf.WriteString(colorCyan)
		if writeCallerWithSkip(buf, 5) {
			buf.WriteString(colorReset)
			buf.WriteByte(' ')
		} else {
			buf.WriteString(colorReset)
		}
	}

	// 从 context 中提取 traceID
	if ctx != nil {
		if traceID, ok := ctx.Value(constant.TraceIDKey).(string); ok && traceID != "" {
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

func (h *StdHandler) getLevelColor(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return colorBlue
	case slog.LevelInfo:
		return colorCyan
	case slog.LevelWarn:
		return colorYellow
	case slog.LevelError:
		return colorRed
	default:
		return colorReset
	}
}

func (h *StdHandler) appendAttr(buf *bytes.Buffer, attr slog.Attr) {
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

func (h *StdHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	newAttrs = append(newAttrs, h.attrs...)
	newAttrs = append(newAttrs, attrs...)

	return &StdHandler{
		w:     h.w,
		level: h.level,
		attrs: newAttrs,
		group: h.group,
	}
}

func (h *StdHandler) WithGroup(name string) slog.Handler {
	newGroup := name
	if h.group != "" {
		newGroup = h.group + "." + name
	}

	return &StdHandler{
		w:     h.w,
		level: h.level,
		attrs: h.attrs,
		group: newGroup,
	}
}
