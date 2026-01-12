package logger

import (
	"bytes"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/Twelveeee/golib/pool"
)

const (
	callerKey = "caller"
	stackKey  = "stack"
)

var (
	pcsPool = sync.Pool{
		New: func() interface{} {
			return &stackPtr{
				pcs: make([]uintptr, 64),
			}
		},
	}
)

type stackPtr struct {
	pcs []uintptr
}

// Stack retrieve call stack
func Stack() slog.Attr {
	return StackWithSkip(3)
}

// StackWithSkip 返回调用栈的Field
func StackWithSkip(skip int) slog.Attr {
	buf := pool.GlobalBytesPool.Get()
	defer pool.GlobalBytesPool.Put(buf)

	stack := pcsPool.Get().(*stackPtr)
	defer pcsPool.Put(stack)

	callStackSize := runtime.Callers(skip, stack.pcs)
	frames := runtime.CallersFrames(stack.pcs[:callStackSize])

	for {
		frame, more := frames.Next()
		buf.WriteString(frame.File)
		buf.WriteByte(':')
		buf.WriteString(strconv.Itoa(frame.Line))
		if !more {
			break
		}
		buf.WriteByte(';')
	}
	return slog.String(stackKey, buf.String())
}

// CallerField 默认的获取调用栈的Field
func CallerField() slog.Attr {
	return CallerFieldWithSkip(2)
}

// CallerFieldWithSkip 获取调用栈
func CallerFieldWithSkip(skip int) slog.Attr {
	return slog.String(callerKey, callerWithSkip(skip+1))
}

// callerWithSkip 获取调用栈的路径
// 如  xxx/xxx/xxx.go:80
func callerWithSkip(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}

	buf := pool.GlobalBytesPool.Get()
	defer pool.GlobalBytesPool.Put(buf)

	buf.WriteString(CallerPathClean(file))
	buf.WriteByte(':')
	buf.WriteString(strconv.Itoa(line))

	return buf.String()
}

// writeCallerWithSkip 将调用栈路径直接写入到 buffer 中，避免字符串分配
// 返回 true 表示成功写入，false 表示获取失败
func writeCallerWithSkip(buf *bytes.Buffer, skip int) bool {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return false
	}

	buf.WriteString(CallerPathClean(file))
	buf.WriteByte(':')
	buf.WriteString(strconv.Itoa(line))

	return true
}

var pathPrefixes = []string{
	"github.com/",
	"gitlab.com/",
	"github/",
	"go.mod/",
}

// CallerPathClean 对caller的文件路径进行精简
// 原始的是完整的路径，比较长，该方法可以将路径变短
var CallerPathClean = callerPathClean

func callerPathClean(file string) string {
	// 尝试匹配常见的代码托管平台路径
	for _, prefix := range pathPrefixes {
		if idx := strings.Index(file, prefix); idx >= 0 {
			return file[idx+len(prefix):]
		}
	}

	// 如果没有匹配到，返回原始路径
	return file
}
