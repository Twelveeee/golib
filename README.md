# golib

[![Go Version](https://img.shields.io/badge/Go-1.24.4+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

一个功能丰富的 Go 语言通用工具库，提供日志、并发任务管理、对象池和实用工具函数等功能。

##  特性

-  **高性能日志系统** - 基于 `log/slog`，支持异步写入和自动轮转
-  **并发任务管理** - 灵活的并发控制和错误处理
-  **对象池** - 减少内存分配，提升性能
-  **泛型工具函数** - 类型安全的 Slice、Map 操作
-  **本地缓存** - 防缓存击穿的高效缓存实现
-  **并发安全** - 完善的 goroutine 管理和 panic 恢复

##  安装

```bash
go get github.com/Twelveeee/golib
```

##  快速开始

### Logger - 日志系统

```go
package main

import (
    "context"
    "log/slog"
    "github.com/Twelveeee/golib/logger"
)

func main() {
    ctx := context.Background()
    
    // 配置日志
    conf := &logger.Config{
        FileName:      "logs/app.log",
        RotateRule:    "1hour",      // 每小时轮转
        MaxFileNum:    48,            // 保留 48 个文件
        BufferSize:    4096,          // 缓冲区大小
        FlushDuration: 1000,          // 1秒刷新一次
        Level:         slog.LevelInfo,
    }
    
    // 创建日志实例
    l, closeFunc, err := logger.NewLogger(ctx, conf)
    if err != nil {
        panic(err)
    }
    defer closeFunc()
    
    // 使用日志
    l.Info("application started", "version", "1.0.0")
    l.Error("error occurred", "error", "something went wrong")
}
```

### GTask - 并发任务管理

```go
package main

import (
    "fmt"
    "github.com/Twelveeee/golib/gtask"
)

func main() {
    g := &gtask.Group{
        Concurrent:    10,    // 最多 10 个并发
        AllowSomeFail: true,  // 允许部分失败
    }
    
    // 添加任务
    for i := 0; i < 100; i++ {
        id := i
        g.Go(func() error {
            // 执行任务
            fmt.Printf("Task %d completed\n", id)
            return nil
        })
    }
    
    // 等待所有任务完成
    successCount, err := g.Wait()
    fmt.Printf("Success: %d, Error: %v\n", successCount, err)
}
```

### Pool - 对象池

```go
package main

import (
    "fmt"
    "github.com/Twelveeee/golib/pool"
)

func main() {
    // 从全局池获取 Buffer
    buf := pool.GlobalBytesPool.Get()
    defer pool.GlobalBytesPool.Put(buf)
    
    buf.WriteString("Hello, ")
    buf.WriteString("World!")
    
    fmt.Println(buf.String())
}
```

### Utils - 工具函数

#### Slice 操作

```go
package main

import (
    "fmt"
    "github.com/Twelveeee/golib/utils"
)

func main() {
    numbers := []int{1, 2, 3, 4, 5}
    
    // Map 转换
    doubled := utils.Map(numbers, func(n int) int {
        return n * 2
    })
    fmt.Println(doubled) // [2 4 6 8 10]
    
    // Filter 过滤
    evens := utils.Filter(numbers, func(n int) bool {
        return n%2 == 0
    })
    fmt.Println(evens) // [2 4]
    
    // Unique 去重
    items := []int{1, 2, 2, 3, 3, 3}
    unique := utils.Unique(items)
    fmt.Println(unique) // [1 2 3]
}
```

#### Map 操作

```go
package main

import (
    "fmt"
    "github.com/Twelveeee/golib/utils"
)

type User struct {
    ID   int
    Name string
}

func main() {
    users := []User{
        {ID: 1, Name: "Alice"},
        {ID: 2, Name: "Bob"},
    }
    
    // 转换为 Map
    userMap := utils.MapByKey(users, func(u User) int {
        return u.ID
    })
    fmt.Println(userMap[1].Name) // Alice
    
    // 提取列
    names := utils.MapColumn(users, func(u User) string {
        return u.Name
    })
    fmt.Println(names) // [Alice Bob]
}
```

#### 本地缓存

```go
package main

import (
    "fmt"
    "time"
    "github.com/Twelveeee/golib/utils"
)

func main() {
    // 创建缓存，过期时间 5 分钟
    cache := utils.NewLocalCache(5 * time.Minute)
    
    // 设置缓存
    cache.Set("user:1", "Alice")
    
    // 获取缓存
    if val, ok := cache.Get("user:1"); ok {
        fmt.Println(val) // Alice
    }
    
    // GetOrSet - 防止缓存击穿
    data, fromCache, err := cache.GetOrSet("user:2", func() (interface{}, error) {
        // 从数据库加载
        return "Bob", nil
    })
    fmt.Printf("Data: %v, FromCache: %v\n", data, fromCache)
}
```

#### 并发工具

```go
package main

import (
    "fmt"
    "github.com/Twelveeee/golib/utils"
)

func main() {
    // 设置全局 panic 处理器
    utils.SetPanicHandler(func(info interface{}) {
        fmt.Printf("Panic recovered: %v\n", info)
    })
    
    // 安全启动 goroutine
    utils.SafeGo(func() {
        // 即使 panic 也会被捕获
        panic("something went wrong")
    })
    
    // 带回调的 goroutine
    utils.CallbackGo(
        func() {
            fmt.Println("Task running")
        },
        func() {
            fmt.Println("Task completed")
        },
    )
}
```

##  模块说明

### Logger

高性能日志系统，主要特性：

-  自定义日志格式
-  异步写入，高性能
-  自动日志轮转（按小时/天）
-  自动清理过期日志
-  支持 TraceID 追踪
-  调用栈信息记录
-  跨平台支持

**配置选项：**

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `FileName` | `string` | 日志文件路径 | 必填 |
| `RotateRule` | `string` | 轮转规则（1hour/1day/no） | 1hour |
| `MaxFileNum` | `int` | 保留文件数量（-1 不清理） | 48 |
| `BufferSize` | `int` | 缓冲队列大小 | 4096 |
| `WriterTimeout` | `int` | 写入超时（毫秒） | 0 |
| `FlushDuration` | `int` | 刷新间隔（毫秒） | 1000 |
| `Level` | `slog.Level` | 日志级别 | - |

### GTask

并发任务管理，主要特性：

-  控制最大并发数
-  支持部分失败容错
-  自动 panic 恢复
-  任务统计

**配置选项：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `Concurrent` | `int` | 最大并发数（0 不限制） |
| `AllowSomeFail` | `bool` | 是否允许部分失败 |

### Pool

对象池，减少内存分配：

-  `bytes.Buffer` 对象池
-  全局共享池 `GlobalBytesPool`
-  自动重置

### Utils

实用工具函数：

**Slice 操作：**
- `ForEach` - 遍历
- `Map` - 映射转换
- `Filter` - 过滤
- `FindIndex` / `FindItem` - 查找
- `Unique` - 去重
- `InArray` - 判断存在
- `Chunk` - 分块
- `Reverse` - 反转

**Map 操作：**
- `MapByKey` - 按键转 Map
- `MapColumn` - 提取列
- `ArrayKeys` - 获取键
- `ArrayValues` - 获取值

**缓存：**
- `LocalCache` - 本地缓存（防击穿）
- `GenerateCacheKey` - 生成缓存键

**并发：**
- `SafeGo` - 安全 goroutine
- `CallbackGo` - 带回调 goroutine
- `SetPanicHandler` - panic 处理器
- `OnceErr` - 只设置一次的错误

##  依赖

```
golang.org/x/sync v0.19.0
```

##  系统要求

- Go 1.24.4 或更高版本
- 支持 macOS、Linux、Windows


##  许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件


---
