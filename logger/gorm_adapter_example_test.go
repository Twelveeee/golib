package logger_test

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/Twelveeee/golib/logger"
	gormlogger "gorm.io/gorm/logger"
)

// ExampleNewGormAdapter 演示如何将 slog.Logger 转换为 GORM logger
func ExampleNewGormAdapter() {
	// 创建一个 slog.Logger
	slogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// 方式1: 使用默认配置
	gormLogger := logger.NewGormAdapter(slogger)
	_ = gormLogger

	// 方式2: 使用自定义配置
	gormLoggerCustom := logger.NewGormAdapter(
		slogger,
		logger.WithGormLogLevel(gormlogger.Info),
		logger.WithSlowThreshold(500*time.Millisecond),
		logger.WithIgnoreRecordNotFoundError(true),
	)
	_ = gormLoggerCustom

	// 在 GORM 中使用
	// db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
	//     Logger: gormLogger,
	// })
}

// ExampleNewGormAdapter_withContext 演示如何使用带上下文的日志记录
func ExampleNewGormAdapter_withContext() {
	ctx := context.Background()

	// 使用 golib 的 logger 包创建 slog.Logger
	conf := &logger.Config{
		FileName:      "./logs/app.log",
		Level:         slog.LevelInfo,
		RotateRule:    "daily",
		MaxFileNum:    7,
		BufferSize:    4096,
		FlushDuration: 1000,
		WriterTimeout: 3000,
	}

	slogger, closeFunc, err := logger.NewLogger(ctx, conf)
	if err != nil {
		panic(err)
	}
	defer closeFunc()

	// 转换为 GORM logger
	gormLogger := logger.NewGormAdapter(
		slogger,
		logger.WithGormLogLevel(gormlogger.Info),
		logger.WithSlowThreshold(200*time.Millisecond),
	)

	// 在 GORM 中使用
	// db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
	//     Logger: gormLogger,
	// })
	_ = gormLogger
}
