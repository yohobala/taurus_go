package field

import (
	"fmt"
	"reflect"
	"time"

	"github.com/yohobala/taurus_go/entity"
	"github.com/yohobala/taurus_go/entity/dialect"
)

type BaseStorage[T any] struct {
	value *T
}

// Set 设置字段的值。
func (b *BaseStorage[T]) Set(value T) error {
	b.value = &value
	return nil
}

// Get 获取字段的值。
func (b *BaseStorage[T]) Get() *T {
	return b.value
}

// Scan 从数据库中读取字段的值。
func (b *BaseStorage[T]) Scan(value interface{}) error {
	if value == nil {
		b.value = nil
		return nil
	}
	return convertAssign(b.value, value)
}

// String 返回字段的字符串表示。
func (b BaseStorage[T]) String() string {
	if b.value == nil {
		return "nil"
	}
	return fmt.Sprintf("%d", *b.value)
}

// Value 返回字段的值，和Get方法不同的是，Value方法返回的是接口类型。
func (i *BaseStorage[T]) Value(dbType dialect.DbDriver) (entity.FieldValue, error) {
	var t T
	return i.toValue(t, dbType)
}

type BaseBuilder[T any] struct {
	desc *entity.Descriptor
}

// Init 初始化字段的描述信息，在代码生成阶段初始化时调用。
//
// Params:
//
//   - desc: 字段的描述信息。
func (b *BaseBuilder[T]) Init(desc *entity.Descriptor) error {
	if b == nil {
		panic("taurus_go/entity field init: nil pointer dereference.")
	}
	b.desc = desc
	return nil
}

// Descriptor 获取字段的描述信息。
func (b *BaseBuilder[T]) Descriptor() *entity.Descriptor {
	return b.desc
}

// AttrType 获取字段的数据库中的类型名，如果返回空字符串，会出现错误。
//
// Params:
//
//   - dbType: 数据库类型。
//
// Returns:
//
//   - 字段的数据库中的类型名。
func (b *BaseBuilder[T]) AttrType(dbType dialect.DbDriver) string {
	var t T
	return attrType(t, dbType)
}

// ValueType 用于设置字段的值在go中类型名称。例如entity.Int64的ValueType为"int64"。
//
// Returns:
//
//   - 字段的值在go中类型名称。
func (b *BaseBuilder[T]) ValueType() string {
	var t T
	return valueType(t)
}

// toValue 将字段的值转换为数据库中的值。
//
// Params:
//
//   - t: 字段的值。
//   - dbType: 数据库类型。
func (b *BaseStorage[T]) toValue(t any, dbType dialect.DbDriver) (entity.FieldValue, error) {
	switch dbType {
	case dialect.PostgreSQL:
		if b.value == nil {
			return nil, nil
		}
		switch t.(type) {
		case int16:
			return *b.value, nil
		case int32:
			return *b.value, nil
		case int64:
			return *b.value, nil
		case []int16:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				return fmt.Sprintf("%d", a), nil
			})
		case []int32:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				return fmt.Sprintf("%d", a), nil
			})
		case []int64:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				return fmt.Sprintf("%d", a), nil
			})
		case [][]int16:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				return fmt.Sprintf("%d", a), nil
			})
		case [][]int32:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				return fmt.Sprintf("%d", a), nil
			})
		case [][]int64:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				return fmt.Sprintf("%d", a), nil
			})
		case [][][]int16:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				return fmt.Sprintf("%d", a), nil
			})
		case [][][]int32:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				return fmt.Sprintf("%d", a), nil
			})
		case [][][]int64:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				return fmt.Sprintf("%d", a), nil
			})
		case [][][][]int16:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				return fmt.Sprintf("%d", a), nil
			})
		case [][][][]int32:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				return fmt.Sprintf("%d", a), nil
			})
		case [][][][]int64:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				return fmt.Sprintf("%d", a), nil
			})
		case [][][][][]int16:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				return fmt.Sprintf("%d", a), nil
			})
		case [][][][][]int32:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				return fmt.Sprintf("%d", a), nil
			})
		case [][][][][]int64:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				return fmt.Sprintf("%d", a), nil
			})
		case bool:
			return *b.value, nil
		case []bool:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				if a.(bool) {
					return "true", nil
				}
				return "false", nil
			})
		case [][]bool:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				if a.(bool) {
					return "true", nil
				}
				return "false", nil
			})
		case [][][]bool:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				if a.(bool) {
					return "true", nil
				}
				return "false", nil
			})
		case [][][][]bool:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				if a.(bool) {
					return "true", nil
				}
				return "false", nil
			})
		case [][][][][]bool:
			return arrayToPGString(*b.value, func(a any) (string, error) {
				if a.(bool) {
					return "true", nil
				}
				return "false", nil
			})
		case string:
			return *b.value, nil
		default:
			return nil, fmt.Errorf("unsupported database type: %v", reflect.TypeOf(t))
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %v", reflect.TypeOf(t))
	}
}

// attrType 返回字段的数据库中的类型名。
func attrType(t any, dbType dialect.DbDriver) string {
	switch dbType {
	case dialect.PostgreSQL:
		switch t.(type) {
		case int16:
			return "int2"
		case int32:
			return "int4"
		case int64:
			return "int8"
		case []int16:
			return "int2[]"
		case []int32:
			return "int4[]"
		case []int64:
			return "int8[]"
		case [][]int16:
			return "int2[][]"
		case [][]int32:
			return "int4[][]"
		case [][]int64:
			return "int8[][]"
		case [][][]int16:
			return "int2[][][]"
		case [][][]int32:
			return "int4[][][]"
		case [][][]int64:
			return "int8[][][]"
		case [][][][]int16:
			return "int2[][][][]"
		case [][][][]int32:
			return "int4[][][][]"
		case [][][][]int64:
			return "int8[][][][]"
		case [][][][][]int16:
			return "int2[][][][][]"
		case [][][][][]int32:
			return "int4[][][][][]"
		case [][][][][]int64:
			return "int8[][][][][]"
		case bool:
			return "boolean"
		case []bool:
			return "boolean[]"
		case [][]bool:
			return "boolean[][]"
		case [][][]bool:
			return "boolean[][][]"
		case [][][][]bool:
			return "boolean[][][][]"
		case [][][][][]bool:
			return "boolean[][][][][]"
		default:
			return ""
		}
	default:
		return ""
	}

}

// valueType 返回字段的值在go中类型名称。
func valueType(t any) string {
	switch t.(type) {
	case int16:
		return "int16"
	case int32:
		return "int32"
	case int64:
		return "int64"
	case []int16:
		return "[]int16"
	case []int32:
		return "[]int32"
	case []int64:
		return "[]int64"
	case [][]int16:
		return "[][]int16"
	case [][]int32:
		return "[][]int32"
	case [][]int64:
		return "[][]int64"
	case [][][]int16:
		return "[][][]int16"
	case [][][]int32:
		return "[][][]int32"
	case [][][]int64:
		return "[][][]int64"
	case [][][][]int16:
		return "[][][][]int16"
	case [][][][]int32:
		return "[][][][]int32"
	case [][][][]int64:
		return "[][][][]int64"
	case [][][][][]int16:
		return "[][][][][]int16"
	case [][][][][]int32:
		return "[][][][][]int32"
	case [][][][][]int64:
		return "[][][][][]int64"
	case bool:
		return "bool"
	case []bool:
		return "[]bool"
	case [][]bool:
		return "[][]bool"
	case [][][]bool:
		return "[][][]bool"
	case [][][][]bool:
		return "[][][][]bool"
	case [][][][][]bool:
		return "[][][][][]bool"
	case string:
		return "string"
	case time.Time:
		return "time.Time"
	case []time.Time:
		return "[]time.Time"
	case [][]time.Time:
		return "[][]time.Time"
	case [][][]time.Time:
		return "[][][]time.Time"
	case [][][][]time.Time:
		return "[][][][]time.Time"
	case [][][][][]time.Time:
		return "[][][][][]time.Time"
	default:
		return ""
	}
}
