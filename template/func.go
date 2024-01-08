package template

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	stringutil "github.com/yohobala/taurus_go/encoding/string"
	"github.com/yohobala/taurus_go/entity/codegen/load"
)

var (
	// Funcs are the predefined template
	// functions used by the codegen.
	Funcs = template.FuncMap{
		"base":               filepath.Base,
		"dict":               dict,
		"toLower":            toLower,
		"toFirstCap":         toFirstCap,
		"toSnakeCase":        stringutil.ToSnakeCase,
		"stringReplace":      strings.Replace,
		"stringHasPrefix":    strings.HasPrefix,
		"sub":                sub,
		"joinStrings":        joinStrings,
		"joinFieldAttrNames": joinFieldAttrNames,
		"joinFieldPrimaies":  joinFieldPrimaies,
	}
	acronyms = make(map[string]struct{})
)

// dict 创建一个map，key/value对的列表。
func dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid dict call")
	}
	dict := make(map[string]interface{})
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

// toLower 把字符串转换为小写。
func toLower(s string) string {
	if _, ok := acronyms[s]; ok {
		return s
	}
	return strings.ToLower(s)
}

func toFirstCap(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
}

func sub(a, b int) int {
	return a - b
}

// joinStrings 把字符串列表连接起来。
func joinStrings(ss ...string) string {
	return strings.Join(ss, "")
}

// joinFieldAttrNames 把字段的AttrName连接起来。
func joinFieldAttrNames(fs []*load.Field) string {
	var ss []string
	for _, f := range fs {
		ss = append(ss, fmt.Sprintf(`'%s'`, f.AttrName))
	}
	return strings.Join(ss, ",")
}

func joinFieldPrimaies(fs []*load.Field) string {
	var fields []*load.Field
	for _, f := range fs {
		if f.Primary > 0 {
			fields = append(fields, f)
		}
	}
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Primary < fields[j].Primary
	})
	var ss []string
	for _, f := range fields {
		ss = append(ss, fmt.Sprintf(`"%s"`, f.AttrName))
	}
	return strings.Join(ss, ",")
}
