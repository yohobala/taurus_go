package load

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/yohobala/taurus_go/cmd"
	"github.com/yohobala/taurus_go/entity"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type (
	// Config 用于从entity package中加载的所有database和entity的配置。
	Config struct {
		// 加载的entity的路径。
		Path string
		// 加载的entity中拥有匿名字段[entity.Entity]的结构体的名称。
		Entities   []string
		BuildFlags []string
		// 加载的entity中拥有匿名字段[entity.Database]的结构体的名称。
		Dbs []DbConfig
	}

	DbConfig struct {
		Name string
		// 存储这个database所拥有的
		Entities EntityMap
	}

	// EntityMap entity的key和类型。
	//
	// 和Config.Entities不同的是，
	// EntityMap是用于记录database中的entity的信息，
	// 而Config.Entities是用于记录entity结构体的名字。
	// 例如：
	// type User struct {
	// 	entity.Database
	// 	User UserEntity
	// }
	// 则这个EntityMap中的内容为：{
	// 	"User": "UserEntity"
	// }
	// 而Config.Entities中的内容为：["UserEntity"]
	EntityMap map[string]string

	// BuilderInfo 是一个用于生成代码的构建器，
	// 包含了entity package的路径和符合条件的database信息。
	BuilderInfo struct {
		// PkgPath 是加载的entity包的Go package路径，之后会传给gen.Config。
		PkgPath string
		// 加载的entity package的模块信息。
		Module *packages.Module
		// entity package的路径符合条件的所有Databases。
		Databases []*Database
	}
)

var (
	// 保存了[entity.Interface]的[reflect.Type]。
	entityInterface = reflect.TypeOf(struct{ entity.EntityInterface }{}).Field(0).Type
	// 保存了[entity.DbInterface]的[reflect.Type]。
	dbInterface = reflect.TypeOf(struct{ entity.DbInterface }{}).Field(0).Type
)

// 加载entity package，并且利用这些信息生成一个Builder。
func (c *Config) Load() (*BuilderInfo, error) {
	// 获取传入路径下的entity信息。
	builder, err := c.load()
	if err != nil {
		return nil, fmt.Errorf("taurus_go/entity: parse entity dir: %w", err)
	}
	if len(c.Entities) == 0 {
		return nil, fmt.Errorf("taurus_go/entity: no entity found in: %s", c.Path)
	}
	// 执行模版。
	var b bytes.Buffer
	err = buildTmpl.ExecuteTemplate(&b, "main", ExeTmplConfig{
		Config:  c,
		Package: builder.PkgPath,
	})
	if err != nil {
		return nil, fmt.Errorf("taurus_go/entity: execute template: %w", err)
	}
	// 格式化生成的代码，并创建目录和文件，最后写入到文件中。
	buf, err := format.Source(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("taurus_go/entity: format template: %w", err)
	}
	if err := os.MkdirAll(".gen", os.ModePerm); err != nil {
		return nil, err
	}
	target := fmt.Sprintf(".gen/%s.go", filename(builder.PkgPath))
	if err := os.WriteFile(target, buf, 0644); err != nil {
		return nil, fmt.Errorf("taurus_go/entity: write file %s: %w", target, err)
	}
	// 清理加载文件。
	defer os.RemoveAll(".gen")
	// 运行生成的代码，解析代码输出，得到entity。
	out, err := cmd.RunGo(target, c.BuildFlags)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(out, "\n") {
		database, err := Unmarshal([]byte(line))
		if err != nil {
			return nil, fmt.Errorf("taurus_go/entity: unmarshal entity %s: %w", line, err)
		}
		builder.Databases = append(builder.Databases, database)
	}
	return builder, nil
}

// 加载传入的路径中符合要求的Entity、database的信息，
// 通过这些信息创建一个Builder。
func (c *Config) load() (*BuilderInfo, error) {
	// 加载指定路径的go包
	// pkgs是一个包的切片，切片的元素数量取决于传入的路径里包含的包的数量
	// 一般来说就2个，一个是c.Path，一个是entityInterface.PkgPath()所在的包。
	pkgs, err := packages.Load(&packages.Config{
		BuildFlags: c.BuildFlags,
		// Load函数需要返回的包的信息
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedModule,
	}, c.Path, entityInterface.PkgPath())
	if err != nil {
		return nil, fmt.Errorf("loading package: %w", err)
	}
	if len(pkgs) < 2 {
		// 检查数量少于2是否是因为 "Go-related"引起的错误
		if err := cmd.List(c.Path, c.BuildFlags); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("missing package information for: %s", c.Path)
	}
	entPkg, loadPkg := pkgs[0], pkgs[1]
	if len(loadPkg.Errors) != 0 {
		return nil, c.loadError(loadPkg.Errors[0])
	}
	if len(entPkg.Errors) != 0 {
		return nil, entPkg.Errors[0]
	}
	// 判断是否是entity接口的包，如果不是翻转。
	if pkgs[0].PkgPath != entityInterface.PkgPath() {
		entPkg, loadPkg = pkgs[1], pkgs[0]
	}
	var names []string
	// 这部分代码是检查，加载的代码中是否有实现了 ent.Interface 接口的结构体。
	// 获取 ent 接口类型：
	iface := entPkg.Types.Scope().Lookup(entityInterface.Name()).Type().Underlying().(*types.Interface)
	var dbs []DbConfig
	dbIface := entPkg.Types.Scope().Lookup(dbInterface.Name()).Type().Underlying().(*types.Interface)
	// 这个循环遍历用户定义的包（loadPkg）中的所有类型定义。
	// loadPkg.TypesInfo.Defs 包含了包中所有类型的定义，其中 k 是定义的标识符（如类型名称），v 是定义本身（如类型信息）。
	for k, v := range loadPkg.TypesInfo.Defs {
		// 这里检查定义 v 是否是一个命名类型（如结构体或接口）。
		typ, ok := v.(*types.TypeName)
		// 如果 v 不是命名类型，或者 k（标识符）不是导出的（即不是公开的），
		// 或者类型 typ没有实现entityInterface接口，则跳过当前迭代。
		if !ok || !k.IsExported() || (!types.Implements(typ.Type(), iface) && !types.Implements(typ.Type(), dbIface)) {
			continue
		}
		// 这里尝试将类型的声明（k.Obj.Decl）断言为 *ast.TypeSpec 类型，这是 Go 语言抽象语法树（AST）中表示类型声明的结构。
		spec, ok := k.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return nil, fmt.Errorf("invalid declaration %T for %s", k.Obj.Decl, k.Name)
		}
		// 这里检查声明的类型（spec.Type）是否是结构体类型。
		structType, ok := spec.Type.(*ast.StructType)
		if !ok {
			return nil, fmt.Errorf("invalid spec type %T for %s", spec.Type, k.Name)
		}

		if types.Implements(typ.Type(), iface) {
			names = append(names, k.Name)
			continue
		}
		if types.Implements(typ.Type(), dbIface) {
			entities := EntityMap{}
			// 遍历结构体的每个字段。
			for _, field := range structType.Fields.List {
				for _, fieldName := range field.Names {
					fieldType := loadPkg.TypesInfo.TypeOf(fieldName)
					if fieldType != nil && types.Implements(fieldType, iface) {
						// 这里可以处理实现了iface接口的字段
						if ident, ok := field.Type.(*ast.Ident); ok {
							entities[fieldName.Name] = ident.Name
						} else {
							return nil, fmt.Errorf("taurus_go/entity invalid field type %T for %s", field.Type, k.Name)
						}

					}
				}
			}
			dbs = append(dbs, DbConfig{
				Name:     k.Name,
				Entities: entities,
			})
			continue
		}

	}
	if len(c.Entities) == 0 {
		c.Entities = names
	}
	if len(c.Dbs) == 0 {
		c.Dbs = dbs
	}

	sort.Strings(c.Entities)
	return &BuilderInfo{PkgPath: loadPkg.PkgPath, Module: loadPkg.Module}, nil
}

func (c *Config) loadError(perr packages.Error) (err error) {
	if strings.Contains(perr.Msg, "import cycle not allowed") {
		if cause := c.cycleCause(); cause != "" {
			perr.Msg += "\n" + cause
		}
	}
	err = perr
	if perr.Pos == "" {
		// Strip "-:" prefix in case of empty position.
		err = errors.New(perr.Msg)
	}
	return err
}

// 检测在给定的 Go 代码包中是否存在可能导致循环依赖的本地类型声明。
func (c *Config) cycleCause() (cause string) {
	// 解析代码目录。
	dir, err := parser.ParseDir(token.NewFileSet(), c.Path, nil, 0)
	// 如果出现解析 错误或无软件包可解析时，忽略报告。
	if err != nil || len(dir) == 0 {
		return
	}
	//查找包含entity的软件包，如果这个操作失败（pkg == nil），则取目录中的第一个包。
	pkg := dir[filepath.Base(c.Path)]
	if pkg == nil {
		for _, v := range dir {
			pkg = v
			break
		}
	}
	// 收集包内的本地类型声明。
	locals := make(map[string]bool)
	for _, f := range pkg.Files {
		for _, d := range f.Decls {
			g, ok := d.(*ast.GenDecl)
			if !ok || g.Tok != token.TYPE {
				continue
			}
			// 遍历包内的所有文件和声明，收集所有公开（exported）的非结构体类型声明。
			// 如果是结构体，遍历结构体的字段来检查是否嵌入了特定的类型。
			for _, s := range g.Specs {
				ts, ok := s.(*ast.TypeSpec)
				if !ok || !ts.Name.IsExported() {
					continue
				}
				// 不是结构体的类型如 "type Role int".
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					locals[ts.Name.Name] = true
					continue
				}
				var embedEntity bool
				astutil.Apply(st.Fields, func(c *astutil.Cursor) bool {
					f, ok := c.Node().(*ast.Field)
					if ok {
						switch x := f.Type.(type) {
						case *ast.SelectorExpr:
							if x.Sel.Name == "Entity" {
								embedEntity = true
							}
						case *ast.Ident:
							if name := strings.ToLower(x.Name); name == "entity" {
								embedEntity = true
							}
						}
					}
					return !embedEntity
				}, nil)
				if !embedEntity {
					locals[ts.Name.Name] = true
				}
			}
		}
	}
	if len(locals) == 0 {
		return
	}
	// 检查 entity 字段中的本地类型使用情况。
	goTypes := make(map[string]bool)
	for _, f := range pkg.Files {
		for _, d := range f.Decls {
			f, ok := d.(*ast.FuncDecl)
			if !ok || f.Name.Name != "Fields" || f.Type.Params.NumFields() != 0 || f.Type.Results.NumFields() != 1 {
				continue
			}
			astutil.Apply(f.Body, func(cursor *astutil.Cursor) bool {
				i, ok := cursor.Node().(*ast.Ident)
				if ok && locals[i.Name] {
					goTypes[i.Name] = true
				}
				return true
			}, nil)
		}
	}
	names := make([]string, 0, len(goTypes))
	for k := range goTypes {
		names = append(names, strconv.Quote(k))
	}
	sort.Strings(names)
	if len(names) > 0 {
		cause = fmt.Sprintf("To resolve this issue, move the custom types used by the generated code to a separate package: %s", strings.Join(names, ", "))
	}
	return
}

func filename(pkg string) string {
	name := strings.ReplaceAll(pkg, "/", "_")
	return fmt.Sprintf("gen_%s_%d", name, time.Now().Unix())
}
