package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	dbClickhouseConf "github.com/hecc-blot/db-clickhouse/config"
	dbClickhouseContract "github.com/hecc-blot/db-clickhouse/contract"
	"github.com/hecc-blot/framework/contract/log"
	"github.com/hecc-blot/framework/util"

	"go.uber.org/zap"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/opentelemetry/tracing"
)

// ClickhouseSvc ClickHouse 数据库服务
type ClickhouseSvc struct {
	ctx context.Context
	db  *gorm.DB
}

// 编译期断言：确保 ClickhouseSvc 实现 IDbClickhouse
var _ dbClickhouseContract.IDbClickhouse = (*ClickhouseSvc)(nil)

// NewClickhouse 创建单个 ClickHouse 实例
func NewClickhouse(config *dbClickhouseConf.ClickhouseConfig, logger log.ILog) (dbClickhouseContract.IDbClickhouse, func(), error) {
	chDb, err := gorm.Open(clickhouse.Open(buildDsn(config)), initGormConfig(logger, config.SlowThreshold))
	if err != nil {
		return nil, func() {}, err
	}

	// 注册 OpenTelemetry 追踪插件，SQL 执行自动生成 span
	useOtelPlugin(chDb)

	sqlDb, err := chDb.DB()
	if err != nil {
		return nil, func() {}, err
	}

	setSqlDbPool(sqlDb, config.MaxIdleConn, config.MaxOpenConn, config.ConnMaxLifetime)

	return &ClickhouseSvc{db: chDb}, func() {
		sqlDb.Close()
	}, nil
}

// buildDsn 组装 ClickHouse DSN
func buildDsn(config *dbClickhouseConf.ClickhouseConfig) string {
	return fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s?dial_timeout=%ds",
		config.Username,
		config.Password,
		config.Ip,
		config.Port,
		config.Database,
		config.ConnectTimeout,
	)
}

// WithContext 设置上下文 — 返回副本，不修改原实例，保证并发安全
func (c *ClickhouseSvc) WithContext(ctx context.Context) dbClickhouseContract.IDbClickhouse {
	ctx = util.ExtractContext(ctx)
	return &ClickhouseSvc{ctx: ctx, db: c.db.WithContext(ctx)}
}

// GetInstance 返回底层 GORM 实例，供执行高级查询
func (c *ClickhouseSvc) GetInstance() any {
	return c.db
}

// Insert 单条写入（追加）
func (c *ClickhouseSvc) Insert(doc interface{}) error {
	return c.db.Create(doc).Error
}

// BatchInsert 批量写入（追加）
func (c *ClickhouseSvc) BatchInsert(docs interface{}) error {
	return c.db.Create(docs).Error
}

// Where 条件 — 返回副本，不修改原实例
func (c *ClickhouseSvc) Where(args ...interface{}) dbClickhouseContract.IDbClickhouse {
	return &ClickhouseSvc{ctx: c.ctx, db: c.db.Where(args[0], args[1:]...)}
}

// Order 排序 — 返回副本，不修改原实例
func (c *ClickhouseSvc) Order(fields ...string) dbClickhouseContract.IDbClickhouse {
	return &ClickhouseSvc{ctx: c.ctx, db: c.db.Order(fields)}
}

// Select 选择字段 — 返回副本，不修改原实例
func (c *ClickhouseSvc) Select(args ...interface{}) dbClickhouseContract.IDbClickhouse {
	return &ClickhouseSvc{ctx: c.ctx, db: c.db.Select(args[0], args[1:]...)}
}

// GroupBy 分组 — 返回副本，不修改原实例
func (c *ClickhouseSvc) GroupBy(fields ...string) dbClickhouseContract.IDbClickhouse {
	return &ClickhouseSvc{ctx: c.ctx, db: c.db.Group(strings.Join(fields, ", "))}
}

// Limit 限制 — 返回副本，不修改原实例
func (c *ClickhouseSvc) Limit(v int) dbClickhouseContract.IDbClickhouse {
	return &ClickhouseSvc{ctx: c.ctx, db: c.db.Limit(v)}
}

// Offset 偏移 — 返回副本，不修改原实例
func (c *ClickhouseSvc) Offset(v int) dbClickhouseContract.IDbClickhouse {
	return &ClickhouseSvc{ctx: c.ctx, db: c.db.Offset(v)}
}

// Count 统计数量
func (c *ClickhouseSvc) Count() (int64, error) {
	var count int64
	err := c.db.Count(&count).Error
	return count, err
}

// Find 查询多条
func (c *ClickhouseSvc) Find(dst interface{}) error {
	return c.db.Find(dst).Error
}

// Take 获取一条
func (c *ClickhouseSvc) Take(dst interface{}) error {
	return c.db.Take(dst).Error
}

// ---- 以下 GORM 公共配置与 db 模块保持一致，独立实现避免跨模块耦合 ----

func initGormConfig(logger log.ILog, slowThreshold int) *gorm.Config {
	return &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   newILogGormLogger(logger, slowThreshold),
	}
}

func useOtelPlugin(db *gorm.DB) {
	db.Use(tracing.NewPlugin(tracing.WithoutMetrics()))
}

func setSqlDbPool(sqlDb *sql.DB, maxIdleConn, maxOpenConn, connMaxLifetime int) {
	sqlDb.SetMaxIdleConns(maxIdleConn)
	sqlDb.SetMaxOpenConns(maxOpenConn)
	sqlDb.SetConnMaxLifetime(time.Second * time.Duration(connMaxLifetime))
}

// iLogGormLogger 是 log.ILog 到 GORM logger.Interface 的适配器
type iLogGormLogger struct {
	logger        log.ILog
	slowThreshold time.Duration
}

func (gl *iLogGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return gl
}

func (gl *iLogGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	gl.logger.Info(ctx, msg, data...)
}

func (gl *iLogGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	gl.logger.Warn(ctx, msg, data...)
}

func (gl *iLogGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	gl.logger.Error(ctx, msg, data...)
}

func (gl *iLogGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	switch {
	case err != nil:
		gl.Error(ctx, "SQL Trace",
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql))
	case elapsed > gl.slowThreshold && gl.slowThreshold > 0:
		gl.Warn(ctx, "Slow SQL",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql))
	default:
		gl.Info(ctx, "SQL Trace",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql))
	}
}

func newILogGormLogger(logger log.ILog, slowThreshold int) logger.Interface {
	return &iLogGormLogger{
		logger:        logger,
		slowThreshold: time.Duration(slowThreshold) * time.Millisecond,
	}
}
