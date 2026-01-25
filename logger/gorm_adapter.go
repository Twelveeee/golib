package logger

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	gormLogger "gorm.io/gorm/logger"
)

// GormAdapter 将 slog.Logger 适配为 gorm.logger.Interface
type GormAdapter struct {
	logger                    *slog.Logger
	logLevel                  gormLogger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

// GormAdapterOption 配置选项
type GormAdapterOption func(*GormAdapter)

// WithGormLogLevel 设置日志级别
func WithGormLogLevel(level gormLogger.LogLevel) GormAdapterOption {
	return func(a *GormAdapter) {
		a.logLevel = level
	}
}

// WithSlowThreshold 设置慢查询阈值
func WithSlowThreshold(threshold time.Duration) GormAdapterOption {
	return func(a *GormAdapter) {
		a.slowThreshold = threshold
	}
}

// WithIgnoreRecordNotFoundError 设置是否忽略 RecordNotFound 错误
func WithIgnoreRecordNotFoundError(ignore bool) GormAdapterOption {
	return func(a *GormAdapter) {
		a.ignoreRecordNotFoundError = ignore
	}
}

// NewGormAdapter 创建一个新的 GORM 日志适配器
func NewGormAdapter(logger *slog.Logger, opts ...GormAdapterOption) gormLogger.Interface {
	adapter := &GormAdapter{
		logger:                    logger,
		logLevel:                  gormLogger.Info,
		slowThreshold:             200 * time.Millisecond,
		ignoreRecordNotFoundError: false,
	}

	for _, opt := range opts {
		opt(adapter)
	}

	return adapter
}

// LogMode 实现 gorm logger.Interface
func (a *GormAdapter) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newAdapter := *a
	newAdapter.logLevel = level
	return &newAdapter
}

// Info 实现 gorm logger.Interface
func (a *GormAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	if a.logLevel >= gormLogger.Info {
		a.logWithoutCaller(ctx, slog.LevelInfo, fmt.Sprintf(msg, data...))
	}
}

// Warn 实现 gorm logger.Interface
func (a *GormAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if a.logLevel >= gormLogger.Warn {
		a.logWithoutCaller(ctx, slog.LevelWarn, fmt.Sprintf(msg, data...))
	}
}

// Error 实现 gorm logger.Interface
func (a *GormAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	if a.logLevel >= gormLogger.Error {
		a.logWithoutCaller(ctx, slog.LevelError, fmt.Sprintf(msg, data...))
	}
}

// Trace 实现 gorm logger.Interface，用于记录 SQL 执行信息
func (a *GormAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if a.logLevel <= gormLogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 清理 SQL 中的换行符和多余空格
	sql = cleanSQL(sql)

	switch {
	case err != nil && a.logLevel >= gormLogger.Error && (!errors.Is(err, gormLogger.ErrRecordNotFound) || !a.ignoreRecordNotFoundError):
		// 记录错误
		a.logAttrsWithoutCaller(ctx, slog.LevelError, "gorm trace error",
			slog.String("sql", sql),
			slog.Int64("rows", rows),
			slog.Duration("elapsed", elapsed),
			slog.String("error", err.Error()),
		)
	case elapsed > a.slowThreshold && a.slowThreshold != 0 && a.logLevel >= gormLogger.Warn:
		// 记录慢查询
		a.logAttrsWithoutCaller(ctx, slog.LevelWarn, "gorm slow query",
			slog.String("sql", sql),
			slog.Int64("rows", rows),
			slog.Duration("elapsed", elapsed),
			slog.Duration("threshold", a.slowThreshold),
		)
	case a.logLevel >= gormLogger.Info:
		// 记录普通查询
		a.logAttrsWithoutCaller(ctx, slog.LevelInfo, "gorm trace",
			slog.String("sql", sql),
			slog.Int64("rows", rows),
			slog.Duration("elapsed", elapsed),
		)
	}
}

// cleanSQL 清理 SQL 语句中的换行符和多余空格
func cleanSQL(sql string) string {
	// 替换所有换行符为空格
	sql = strings.ReplaceAll(sql, "\n", " ")
	sql = strings.ReplaceAll(sql, "\r", " ")
	sql = strings.ReplaceAll(sql, "\t", " ")

	// 压缩多个连续空格为单个空格
	for strings.Contains(sql, "  ") {
		sql = strings.ReplaceAll(sql, "  ", " ")
	}

	// 去除首尾空格
	return strings.TrimSpace(sql)
}

// logWithoutCaller 记录日志但不包含 caller 信息
func (a *GormAdapter) logWithoutCaller(ctx context.Context, level slog.Level, msg string) {
	if !a.logger.Enabled(ctx, level) {
		return
	}
	r := slog.NewRecord(time.Now(), level, msg, 0)
	_ = a.logger.Handler().Handle(ctx, r)
}

// logAttrsWithoutCaller 记录带属性的日志但不包含 caller 信息
func (a *GormAdapter) logAttrsWithoutCaller(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	if !a.logger.Enabled(ctx, level) {
		return
	}
	r := slog.NewRecord(time.Now(), level, msg, 0)
	r.AddAttrs(attrs...)
	_ = a.logger.Handler().Handle(ctx, r)
}
