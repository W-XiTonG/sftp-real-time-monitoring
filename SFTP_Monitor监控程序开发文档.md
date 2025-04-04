

---

# SFTP 监控程序开发文档

## 一、项目目录结构说明

```bash
sftp-monitor/
├── cmd/                     # 命令行入口目录
│   └── sftp-monitor/        # 主程序包
│       └── main.go          # 程序入口（初始化、信号处理、主循环）
├── internal/                # 内部实现模块（不对外暴露）
│   ├── config/              # 配置管理
│   │   └── config.go        # 配置加载、验证、环境变量绑定
│   ├── sftp/                # SFTP核心连接模块
│   │   └── client.go        # 客户端封装（连接池、重连、状态管理）
│   ├── monitor/             # 监控核心逻辑
│   │   └── watcher.go       # 文件监控器（差异检测、事件触发）
│   ├── handler/             # 事件处理模块
│   │   └── handler.go       # 处理接口定义与基础实现
│   └── utils/               # 通用工具包
│       └── utils.go         # 日志、加密、文件操作等工具函数
├── configs/                 # 配置示例文件
│   └── config.yaml          # YAML配置模板（含完整注释）
├── pkg/                     # 可复用公共库
│   └── filecache/           # 文件状态缓存系统
│       ├── cache.go         # 缓存数据结构与持久化
│       └── serializer.go    # 缓存序列化/反序列化
├── go.mod                   # Go模块定义
├── go.sum                   # 依赖校验文件
└── README.md                # 项目说明文档
```

---

## 二、目录结构详解

### 1. `cmd/sftp-monitor/`
- **定位**：程序入口层
- **核心文件**：`main.go`
- **职责**：
  - 命令行参数解析
  - 依赖组件初始化（配置、日志、客户端）
  - 信号监听（SIGTERM/SIGINT）
  - 优雅关闭流程控制
- **典型代码**：
  ```go
  func init() {
      rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", 
          "config.yaml", "配置文件路径")
  }
  
  func main() {
      // 初始化上下文
      ctx, cancel := context.WithCancel(context.Background())
      defer cancel()
  
      // 启动监控协程
      go watcher.Run(ctx)
  
      // 等待终止信号
      <-utils.WaitForTermSignal()
      watcher.Stop()
  }
  ```

---

### 2. `internal/config/`
- **定位**：配置管理中心
- **核心文件**：`config.go`
- **功能特性**：
  - 多格式配置支持（YAML/JSON/ENV）
  - 自动类型转换（时间单位、大小单位）
  - 配置版本兼容性检查
  - 敏感字段自动脱敏
- **配置加载流程**：
  ```
  读取文件 → 环境变量覆盖 → 字段验证 → 密钥解密 → 生成最终配置
  ```
- **安全设计**：
  ```go
  func (c *SFTPConfig) DecryptFields() error {
      if c.EncryptedPassword != "" {
          decrypted, err := utils.AESDecrypt(c.EncryptedPassword, masterKey)
          c.Password = decrypted
      }
  }
  ```

---

### 3. `internal/sftp/`
- **定位**：SFTP连接管理层
- **核心文件**：`client.go`
- **核心组件**：
  - **连接池**：维护多个活跃连接
    ```go
    type ConnectionPool struct {
        connections chan *sftp.Client
        factory     func() (*sftp.Client, error)
    }
    ```
  - **心跳检测**：定期验证连接有效性
    ```go
    func (p *Pool) StartHeartbeat(interval time.Duration) {
        go func() {
            ticker := time.NewTicker(interval)
            for {
                <-ticker.C
                p.CheckAliveConnections()
            }
        }()
    }
    ```
  - **传输监控**：记录流量统计
    ```go
    type TrafficStats struct {
        TotalRead  int64 `json:"total_read"`  // 总读取字节
        TotalWrite int64 `json:"total_write"` // 总写入字节
    }
    ```

---

### 4. `internal/monitor/`
- **定位**：监控系统核心
- **核心文件**：`watcher.go`
- **监控策略**：
  - **轮询模式**：定时全量扫描（默认）
  - **事件模式**：服务器事件通知（需服务器支持）
- **关键技术点**：
  ```go
  // 差异检测算法
  func (w *Watcher) compareFiles(current FileSnapshot) []FileEvent {
      events := make([]FileEvent, 0)
  
      // 使用位掩码快速判断变化类型
      const (
          CREATED = 1 << iota
          MODIFIED
          DELETED
      )
  
      // 并行比对加速
      var wg sync.WaitGroup
      resultChan := make(chan FileEvent, 100)
  
      // 检查新增/修改文件
      for name, newMeta := range current {
          oldMeta, exists := w.cache.Get(name)
          wg.Add(1)
          go func() {
              defer wg.Done()
              if !exists {
                  resultChan <- FileEvent{Type: CREATED, File: name}
              } else if newMeta.Hash != oldMeta.Hash {
                  resultChan <- FileEvent{Type: MODIFIED, File: name}
              }
          }()
      }
  
      // 检查删除文件
      for oldName := range w.cache.GetAll() {
          if _, exists := current[oldName]; !exists {
              resultChan <- FileEvent{Type: DELETED, File: oldName}
          }
      }
  
      go func() {
          wg.Wait()
          close(resultChan)
      }()
  
      for event := range resultChan {
          events = append(events, event)
      }
      return events
  }
  ```

---

### 5. `internal/handler/`
- **定位**：事件处理抽象层
- **核心文件**：`handler.go`
- **设计模式**：
  - **观察者模式**：支持多个处理程序并行处理
  - **责任链模式**：事件可被多个处理器依次处理
- **扩展接口**：
  ```go
  // 事件处理器接口
  type EventHandler interface {
      HandleCreated(file string, meta FileMeta) error
      HandleModified(file string, oldMeta, newMeta FileMeta) error
      HandleDeleted(file string, meta FileMeta) error
  }
  
  // Webhook处理器示例
  type WebhookHandler struct {
      endpoint string
      timeout  time.Duration
  }
  
  func (h *WebhookHandler) HandleCreated(file string, meta FileMeta) error {
      payload := map[string]interface{}{
          "event":    "created",
          "file":     file,
          "size":     meta.Size,
          "modified": meta.ModTime.Unix(),
      }
      return h.sendJSON(payload)
  }
  ```

---

### 6. `pkg/filecache/`
- **定位**：状态持久化模块
- **核心特性**：
  - **缓存淘汰策略**：LRU + TTL 双机制
  - **持久化存储**：定期保存到磁盘
  - **版本控制**：自动迁移旧版缓存格式
- **数据结构**：
  ```go
  type FileCache struct {
      sync.RWMutex
      store       map[string]CacheEntry
      lruList     *list.List               // LRU队列
      ttl         time.Duration            // 默认存活时间
      persistPath string                   // 持久化文件路径
  }
  
  type CacheEntry struct {
      Key        string
      Hash       string
      Size       int64
      ModTime    time.Time
      ExpireTime time.Time
      lruNode    *list.Element            // LRU链表节点
  }
  ```
- **缓存恢复流程**：
  ```
  启动时检查持久化文件 → 反序列化数据 → 验证时间戳 → 重建内存结构
  ```

---

## 三、扩展开发指南

### 1. 添加新配置项步骤：
1. 在 `internal/config/config.go` 的结构体中添加字段
2. 实现字段验证逻辑（如需要）
3. 更新 `configs/config.yaml` 示例文件
4. 在相关模块中使用新配置项

### 2. 开发新处理程序：
1. 在 `internal/handler/` 下新建 `xxx_handler.go`
2. 实现 `EventHandler` 接口
3. 在 `main.go` 中注册处理器：
   ```go
   watcher.RegisterHandler(&handler.NewHandler{})
   ```

### 3. 性能优化建议：
- **并发扫描**：将目录遍历任务分解为多个并发子任务
- **增量哈希**：对大文件只计算头部+中部+尾部的哈希组合
- **缓存预热**：启动时预加载历史缓存文件

---

