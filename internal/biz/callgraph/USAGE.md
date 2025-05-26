# CallGraph 使用指南

## 简化后的API设计

### ✅ 推荐使用方式：Execute方法

```go
package main

import (
    "context"
    "log"
    
    "github.com/toheart/goanalysis/internal/biz/callgraph"
)

func main() {
    // 创建程序分析器
    analyzer := callgraph.NewProgramAnalysis(
        "/path/to/your/project",
        logger,
        dataStore,
        callgraph.WithAlgo(callgraph.CallGraphTypeVta),
        callgraph.WithIgnorePaths("vendor,test"),
    )
    
    // 一键执行：分析 + 构建树 + 保存数据库（并发安全）
    statusChan := make(chan []byte, 100)
    ctx := context.Background()
    
    // 监听状态更新（可选）
    go func() {
        for status := range statusChan {
            log.Printf("Status: %s", string(status))
        }
    }()
    
    // 执行完整分析流程（内部自动处理并发）
    if err := analyzer.Execute(ctx, statusChan); err != nil {
        log.Fatalf("Analysis failed: %v", err)
    }
    
    log.Println("Analysis completed successfully!")
}
```

### 🔧 高级使用（分步执行）

如果需要更细粒度的控制，仍然可以分步执行：

```go
// 步骤1：构建调用图树
if err := analyzer.SetTree(statusChan); err != nil {
    log.Fatal(err)
}

// 步骤2：保存到数据库
if err := analyzer.SaveData(ctx, statusChan); err != nil {
    log.Fatal(err)
}
```

## 并发安全设计

### 🚨 之前的问题
```go
// 串行执行导致的死锁问题
func Execute() {
    setTree()  // 生产数据到channel（生产者）
    saveData() // 消费channel数据（消费者）
}
// 问题：当channel缓冲区满时，setTree会阻塞等待消费者
```

### ✅ 现在的解决方案
```go
// 并发执行，生产者和消费者同时工作
func Execute() {
    // 启动消费者goroutine
    go consumeData()
    
    // 主线程生产数据
    produceData()
    
    // 关闭channels
    close(channels)
    
    // 等待消费者完成
    wait()
}
```

### 🔄 执行流程

1. **初始化阶段**: 创建数据库表
2. **并发阶段**: 
   - 启动消费者goroutine（`consumeData`）
   - 主线程生产数据（`produceData`）
3. **同步阶段**: 
   - 关闭channels通知消费者
   - 等待消费者处理完成

## API对比

### ❌ 之前：串行执行 + 死锁风险
```go
// 用户需要了解内部执行顺序，且存在死锁风险
analyzer := NewProgramAnalysis(...)
analyzer.SetTree(statusChan)        // 可能阻塞
analyzer.SaveData(ctx, statusChan)  // 消费数据
```

### ✅ 现在：并发安全 + 简单易用
```go
// 用户只需调用一个方法，内部自动处理并发
analyzer := NewProgramAnalysis(...)
analyzer.Execute(ctx, statusChan)   // 并发安全，无死锁
```

## 核心改进

1. **并发安全**：解决了生产者-消费者死锁问题
2. **性能提升**：生产和消费并发进行，提高效率
3. **封装性**：隐藏了复杂的并发控制逻辑
4. **易用性**：用户无需了解内部并发细节
5. **向后兼容**：保留了原有的公开方法

## 内部架构

```
Execute()
├── 初始化数据库表
├── 启动消费者goroutine ──┐
│   ├── 消费节点数据      │
│   └── 消费边数据        │  并发执行
├── 生产调用图数据 ──────┘
├── 关闭channels
└── 等待消费者完成
```

## 配置选项

支持以下配置选项：

- `WithAlgo(algo)`: 设置分析算法（static/cha/rta/vta）
- `WithIgnorePaths(paths)`: 设置忽略路径
- `WithOnlyPkg(pkg)`: 只分析特定包
- `WithOutputDir(dir)`: 设置输出目录
- `WithCacheDir(dir)`: 设置缓存目录
- `WithCacheFlag(flag)`: 是否使用缓存 