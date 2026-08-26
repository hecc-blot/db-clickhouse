# hecc-blot-db-clickhouse

基于 GORM 的 ClickHouse 分析型数据库组件：链式查询、批量追加写入，SQL 自动接入链路追踪。无事务、无更新删除（ClickHouse 为 append-only 存储）。

## 安装

```bash
go get github.com/hecc-blot/db-clickhouse
```

## 接口定义

```go
import (
    dbContract "github.com/hecc-blot/db/contract"
    dbClickhouseContract "github.com/hecc-blot/db-clickhouse/contract"
)

type IDbClickhouse interface {
    dbContract.IDbBase
    WithContext(ctx context.Context) IDbClickhouse
    Insert(doc interface{}) error
    BatchInsert(docs interface{}) error
    Where(args ...interface{}) IDbClickhouse
    Order(fields ...string) IDbClickhouse
    Select(args ...interface{}) IDbClickhouse
    GroupBy(fields ...string) IDbClickhouse
    Limit(v int) IDbClickhouse
    Offset(v int) IDbClickhouse
    Count() (int64, error)
    Find(dst interface{}) error
    Take(dst interface{}) error
}
```

## 初始化

```go
import (
    dbClickhouse "github.com/hecc-blot/db-clickhouse/service"
)

chDb, clearUp, err := dbClickhouse.NewClickhouse(&config.Clickhouse, logSvc)
if err != nil {
    panic(err)
}
defer clearUp()

container.Set(new(dbClickhouseContract.IDbClickhouse), chDb)
```

业务方直接注入 `IDbClickhouse`，每个请求用 `WithContext(ctx)` 取副本：

```go
type LogStatApi struct {
    Ch dbClickhouseContract.IDbClickhouse `inject:""`
}

func (a LogStatApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
    db := a.Ch.WithContext(ctx)   // 返回绑定请求上下文的副本，并发安全
    var rows []LogModel
    if err := db.
        Select("date, count(*) as cnt").
        Where("date >= ?", "2026-01-01").
        GroupBy("date").
        Order("date DESC").
        Find(&rows); err != nil {
        return nil, errorSvc.NewError(response.Fail, err)
    }
    return rows, nil
}
```

## CRUD 操作

```go
// Insert — 单条追加
err := chDb.Insert(&LogModel{Date: "2026-08-26", Msg: "hello"})

// BatchInsert — 批量追加
err := chDb.BatchInsert(&[]LogModel{{Date: "2026-08-26"}, {Date: "2026-08-27"}})

// Find — 查询多条
err := chDb.Where("date >= ?", "2026-08-01").Limit(100).Find(&rows)

// Count — 统计
count, err := chDb.Where("date >= ?", "2026-08-01").Count()

// 分组聚合
err = chDb.Select("date, count(*) as cnt").GroupBy("date").Order("date").Find(&rows)
```

## 配置

```yaml
clickhouse:
  ip: 127.0.0.1
  port: 9000               # 原生协议端口
  username: default
  password: ""
  database: logs
  connect_timeout: 3       # 建连超时（秒）
  max_idle_conn: 10
  max_open_conn: 100
  conn_max_lifetime: 3600  # 连接最大生命周期（秒）
  slow_threshold: 200      # 慢查询阈值（毫秒），0 不记录
```

## SQL 链路追踪

与 db 模块一致，通过 GORM OpenTelemetry 插件自动生成 span，只依赖第三方 otel，不依赖 trace 模块。初始化顺序要求：先初始化 trace，再初始化 db-clickhouse。

## 相关模块

| 模块 | 说明 |
|------|------|
| [db](https://github.com/hecc-blot/db) | 关系型（MySQL / PostgreSQL），提供 `IDbBase` |
| [db-mongo](https://github.com/hecc-blot/db-mongo) | 文档型（MongoDB） |
| [db-es](https://github.com/hecc-blot/db-es) | 搜索型（Elasticsearch） |
