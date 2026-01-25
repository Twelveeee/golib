package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Twelveeee/golib/logger/handler"
	"github.com/Twelveeee/golib/logger/writer"
)

func NewLogger(ctx context.Context, conf *Config) (l *slog.Logger, closeFunc func() error, errResult error) {
	// 验证和设置默认值
	if err := conf.Validate(); err != nil {
		return nil, nil, fmt.Errorf("invalid config: %w", err)
	}
	conf.SetDefaults()

	closeFns := make([]func() error, 0, 6)
	var closeOnce sync.Once
	var closeErr error

	closeWritersFunc := func() error {
		closeOnce.Do(func() {
			var builder strings.Builder
			for idx, fn := range closeFns {
				if e := fn(); e != nil {
					builder.WriteString(fmt.Sprintf("idx=%d error=%s;", idx, e))
				}
			}
			if builder.Len() > 0 {
				closeErr = fmt.Errorf("logger close with errors: %s", builder.String())
			}
		})
		return closeErr
	}

	writer, err := conf.getWriter()
	if err != nil {
		return nil, nil, fmt.Errorf("init logger (%q) failed: %w", conf.FileName, err)
	}

	closeFns = append(closeFns, writer.Close)

	// 如果是 Debug 级别，同时输出到标准输出
	var logHandler slog.Handler
	if conf.Level == slog.LevelDebug {
		fileHandler := handler.NewDefaultHandler(writer, conf.Level)
		stdoutHandler := handler.NewStdHandler(os.Stdout, conf.Level)
		logHandler = handler.NewMultiHandler(fileHandler, stdoutHandler)
	} else {
		logHandler = handler.NewDefaultHandler(writer, conf.Level)
	}

	l = slog.New(logHandler)

	if ctx != nil {
		go func() {
			<-ctx.Done()
			if e := closeWritersFunc(); e != nil {
				fmt.Fprintf(os.Stderr, "%s logger shutdown error: %v\n", time.Now(), e)
			}
		}()
	}

	return l, closeWritersFunc, nil
}

func (conf *Config) getWriter() (io.WriteCloser, error) {
	if conf.writer != nil {
		return conf.writer, nil
	}
	// 以下内容是创建一个writer所需要的配置
	rp, err := writer.NewSimpleRotateProducer(conf.RotateRule, conf.FileName)
	if err != nil {
		return nil, err
	}

	writerOption := &writer.RotateOption{
		FileProducer:  rp,
		FlushDuration: time.Duration(conf.FlushDuration) * time.Millisecond,
		CheckDuration: 1 * time.Second,
		MaxFileNum:    conf.MaxFileNum,
	}

	w, errRw := writer.NewRotate(writerOption)
	if errRw != nil {
		return nil, errRw
	}

	awc := writer.NewAsync(conf.BufferSize, time.Millisecond*time.Duration(conf.WriterTimeout), w)
	return awc, nil
}
