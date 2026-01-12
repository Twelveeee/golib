package timer

import (
	"time"
)

// 将当前时间函数替换为可配置的函数，方便测试
var nowFunc = time.Now
