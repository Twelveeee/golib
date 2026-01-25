package logger

import (
	"errors"
	"io"
	"log/slog"
)

type Config struct {
	// 日志文件名
	// 如  log/service/service.log
	FileName string `json:"fileName" yaml:"fileName"`

	// 文件切分规则，如 1hour,1day,no,默认为1hour
	RotateRule string `json:"rotateRule" yaml:"rotateRule"`

	// 保留最多日志文件数，默认为48，若为-1,则不会清理
	// 对于 FileName 所在目录下的 以FileName为前缀的文件将自动进行清理
	// 清理后剩余文件数量，清理周期同 RotateRule
	MaxFileNum int `json:"maxFileNum" yaml:"maxFileNum"`

	// 日志内容待写缓冲队列大小
	// 若<0, 则是同步的
	// 若为0，则使用默认值4096
	BufferSize int `json:"bufferSize" yaml:"bufferSize"`

	// 日志进入待写队列超时时间，毫秒
	// 默认为0，不超时，若出现落盘慢的时候，调用写日志的地方会出现同步等待
	WriterTimeout int `json:"writerTimeout" yaml:"writerTimeout"`

	// 日志落盘刷新间隔，毫秒
	// 若<=0，使用默认值1000
	FlushDuration int `json:"flushDuration" yaml:"flushDuration"`

	// 日志等级
	Level slog.Level `json:"level" yaml:"level"`

	writer io.WriteCloser
}

// Validate 验证配置是否有效
func (c *Config) Validate() error {
	if c.FileName == "" {
		return errors.New("FileName is required")
	}
	return nil
}

// SetDefaults 设置默认值
func (c *Config) SetDefaults() {
	if c.RotateRule == "" {
		c.RotateRule = "1hour"
	}
	if c.MaxFileNum == 0 {
		c.MaxFileNum = 48
	}
	if c.BufferSize == 0 {
		c.BufferSize = 4096
	}
	if c.FlushDuration <= 0 {
		c.FlushDuration = 1000
	}
}
