package logger

// ContextKey 用于从 context 中提取值的 key 类型
type ContextKey string

const (
	// TraceIDKey context 中 traceID 的 key
	TraceIDKey ContextKey = "traceID"
)
