package writer

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Twelveeee/golib/timer"
)

// RotateInfo 文件信息
type RotateInfo struct {
	// 原始的名称，相对于FilePath来说就是未添加后缀前的名称
	// 如 xxx/service.log
	RawName string

	// 文件软连接 xxx/service.log
	// 由 FilePath 创建:
	// ln -s FilePath Symlink
	Symlink string

	// 文件路径，日志直接写入到这个文件
	// 如 xxx/service.log.2020072215
	FilePath string
}

// RotateProducer 日志文件名生成发射器
// 可定时创建新的日志文件名
type RotateProducer interface {
	// 获取当前的信息
	Get() RotateInfo

	// 注册回调
	RegisterCallBack(callBackFunc func(info RotateInfo))

	// 停止
	Stop() error
}

// Equal 是否相等
func (info RotateInfo) Equal(other RotateInfo) bool {
	return info.Symlink == other.Symlink && info.FilePath == other.FilePath && info.RawName == other.RawName
}

// NeedSymlink 是否需要软连接
func (info RotateInfo) NeedSymlink() bool {
	return info.Symlink != "" && info.Symlink != info.FilePath
}

// NewRotateProducer 创建一个新的日志切割分发器
func NewRotateProducer(duration time.Duration, producer func() RotateInfo) RotateProducer {
	return &rotateProducer{
		p: timer.NewProducer(duration, func() interface{} {
			return producer()
		}),
	}
}

// NewSimpleRotateProducer 使用已有规则生成具有自动定时变化文件名的分发器
func NewSimpleRotateProducer(rule string, fileNamePrefix string) (RotateProducer, error) {
	if fileNamePrefix == "" {
		return nil, fmt.Errorf("fileNamePrefix is empty")
	}
	rt, has := defaultRotateRules[rule]
	if !has {
		return nil, fmt.Errorf("rule=%q not supported yet", rule)
	}

	return NewRotateProducer(rt.Duration, func() RotateInfo {
		return RotateInfo{
			RawName: fileNamePrefix,
			Symlink: fileNamePrefix,
			FilePath: strings.Join([]string{
				fileNamePrefix,
				rt.SuffixProducer(),
			}, ""),
		}
	}), nil
}

type rotateProducer struct {
	p timer.Producer
}

func (r *rotateProducer) Get() RotateInfo {
	return r.p.Get().(RotateInfo)
}

func (r *rotateProducer) RegisterCallBack(callBackFunc func(info RotateInfo)) {
	r.p.RegisterCallBack(func(value interface{}) {
		callBackFunc(value.(RotateInfo))
	})
}

func (r *rotateProducer) Stop() error {
	r.p.Stop()
	return nil
}

var _ RotateProducer = (*rotateProducer)(nil)

type rotateRule struct {
	Duration       time.Duration
	SuffixProducer func() string
}

var defaultRotateRules = map[string]*rotateRule{
	"1hour": {
		Duration: 1 * time.Hour,
		// 小时  后缀如  .2020072217
		SuffixProducer: func() string {
			return "." + nowFunc().Format("2006010215")
		},
	},
	"1day": {
		Duration: 24 * time.Hour,
		// 天  后缀如  .20200722
		SuffixProducer: func() string {
			return "." + nowFunc().Format("20060102")
		},
	},
	"no": {
		Duration: 0,
		// 无后缀
		SuffixProducer: func() string {
			return ""
		},
	},
	"1min": {
		Duration: 1 * time.Minute,
		// 1分钟 后缀如  .202007221700  .202007221701  .202007221702  .202007221759
		SuffixProducer: func() string {
			now := nowFunc()
			return "." + now.Format("2006010215") + fmt.Sprintf("%02d", now.Minute())
		},
	},
	"5min": {
		Duration: 5 * time.Minute,
		// 5分钟 后缀如  .202007221700  .202007221705  .202007221710  .202007221715
		SuffixProducer: func() string {
			now := nowFunc()
			return "." + now.Format("2006010215") + fmt.Sprintf("%02d", now.Minute()/5*5)
		},
	},
	"10min": {
		Duration: 10 * time.Minute,
		// 10分钟 后缀如  .202007221700  .202007221710  .202007221720  .202007221750
		SuffixProducer: func() string {
			now := nowFunc()
			return "." + now.Format("2006010215") + fmt.Sprintf("%02d", now.Minute()/10*10)
		},
	},
	"15min": {
		Duration: 15 * time.Minute,
		// 15分钟 后缀如  .202007221700  .202007221715  .202007221730  .202007221745
		SuffixProducer: func() string {
			now := nowFunc()
			return "." + now.Format("2006010215") + fmt.Sprintf("%02d", now.Minute()/15*15)
		},
	},
	"30min": {
		Duration: 30 * time.Minute,
		// 30分钟 后缀如  .202007221700  .202007221730
		SuffixProducer: func() string {
			now := nowFunc()
			return "." + now.Format("2006010215") + fmt.Sprintf("%02d", now.Minute()/30*30)
		},
	},
	"": {
		Duration: 1 * time.Hour,
		// 小时  后缀如  .2020072217
		SuffixProducer: func() string {
			return "." + nowFunc().Format("2006010215")
		},
	},
}

// RegisterRotateRule 注册新的文件切分规则
//
// 已内置的规则：
//
//	1hour -> 1小时  后缀如  .2020072217
//	1day  -> 1天    后缀如  .20200722
//	no    -> 无后缀
//	1min  -> 1分钟  后缀如  .202007221700  .202007221701  .202007221702  .202007221759
//	5min  -> 5分钟  后缀如  .202007221700  .202007221705  .202007221710  .202007221715
//	10min -> 10分钟 后缀如  .202007221700  .202007221710  .202007221720  .202007221750
//	30min -> 30分钟 后缀如  .202007221700  .202007221730
//
//	若当前时间是 2020年07月22日 17点34
//	如选择规则 "1hour", 内容会输出到 xxx.2020072217
//	如选择规则 "5min",  内容会输出到 xxx.202007221730
//	如选择规则 "30min", 内容会输出到 xxx.202007221730
//
//	为了让日志清理以一个比较简单的方式能找到所需要清理的文件
//	请保持后缀是一个简单的数字格式，如 .20200722，
//	而不是 .abcd231212 或者 .abc.12345  (可能清理不掉)
func RegisterRotateRule(rule string, duration time.Duration, suffix func() string) error {
	if _, has := defaultRotateRules[rule]; has {
		return errors.New("rule already exists")
	}
	defaultRotateRules[rule] = &rotateRule{
		Duration:       duration,
		SuffixProducer: suffix,
	}
	return nil
}
