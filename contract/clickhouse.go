package contract

import (
	"context"

	dbContract "github.com/hecc-blot/db/contract"
)

// IDbClickhouse ClickHouse 分析型数据库接口（无事务、无更新删除）。
// 复用 db 模块的 IDbBase（GetInstance），自身提供查询 + 追加写入能力。
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
