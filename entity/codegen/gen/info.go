package gen

import (
	"strings"

	stringutil "github.com/yohobala/taurus_go/encoding/string"
	"github.com/yohobala/taurus_go/entity/codegen/load"
)

type (
	// DatabaseInfo 表示一个Builder中的一个数据库节点的信息
	DatabaseInfo struct {
		*Config
		Database *load.Database
	}

	// EntityInfo 表示一个Builder中的一个实体节点的信息
	EntityInfo struct {
		*Config
		Entity *load.Entity
	}
)

// NewDatabaseInfo 初始化一个DatabaseInfo
//
// Params:
//
//   - c: 代码生成的配置。
//   - database: 从entity package中加载的数据库。
//
// Returns:
//
//	0: 数据库信息。
//	1: 错误信息。
func NewDatabaseInfo(c *Config, database *load.Database) (*DatabaseInfo, error) {
	typ := &DatabaseInfo{
		Config:   c,
		Database: database,
	}
	return typ, nil
}

// NewEntityInfo 初始化一个EntityInfo
//
// Params:
//
//   - c: 代码生成的配置。
//   - entity: 从entity package中加载的实体。
//
// Returns:
//
//	0: 实体信息。
//	1: 错误信息。
func NewEntityInfo(c *Config, entity *load.Entity) (*EntityInfo, error) {
	typ := &EntityInfo{
		Config: c,
		Entity: entity,
	}
	return typ, nil
}

// Dir 返回包目录名称
func (t DatabaseInfo) Dir() string {
	return strings.ToLower(t.Database.Name)
}

// Dir 返回包目录名称
func (t EntityInfo) Dir() string {
	return stringutil.ToSnakeCase(t.Entity.AttrName)
}
