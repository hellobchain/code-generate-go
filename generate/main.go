package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	mysqlDsn = flag.String("dns", "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local", "")
	outPath  = flag.String("out", "../demo", "生成代码目录")
	module   = flag.String("mod", "github.com/demo", "go module 名")
)

// ---------- 表/列元数据结构 ----------
type Column struct {
	GoName string // 结构体字段名
	GoType string // go 类型
	Tag    string // gorm tag
	GoTag  string // go tag
}

type Table struct {
	TableName string
	GoName    string
	Columns   []Column
}

// ---------- 主入口 ----------
// 返回连接池 + 所有表结构
func loadTables() (*gorm.DB, []Table) {
	db, err := gorm.Open(mysql.Open(*mysqlDsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect mysql err: %v", err)
	}

	// 1. 拿到当前库名
	var dbName string
	db.Raw("SELECT DATABASE()").Scan(&dbName)

	// 2. 所有表名
	var tableNames []string
	db.Raw(`SELECT table_name
	          FROM information_schema.tables
	          WHERE table_schema = ?
	            AND table_type = 'BASE TABLE'`, dbName).Scan(&tableNames)

	// 3. 逐个表解析
	var tables []Table
	for _, tn := range tableNames {
		cols := loadColumns(db, dbName, tn)
		tables = append(tables, Table{
			TableName: tn,
			GoName:    toGoName(tn),
			Columns:   cols,
		})
	}
	return db, tables
}

// ---------- 列级解析 ----------
func loadColumns(db *gorm.DB, dbName, tableName string) []Column {
	// 查询列信息
	var raw []struct {
		ColumnName    string `gorm:"column:COLUMN_NAME"`
		DataType      string `gorm:"column:DATA_TYPE"`
		IsNullable    string `gorm:"column:IS_NULLABLE"`
		ColumnKey     string `gorm:"column:COLUMN_KEY"`
		ColumnType    string `gorm:"column:COLUMN_TYPE"`    // int(11) / decimal(10,2)
		ColumnComment string `gorm:"column:COLUMN_COMMENT"` // 可后续生成字段注释
	}
	db.Raw(`SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_KEY, COLUMN_TYPE, COLUMN_COMMENT
	          FROM information_schema.columns
	          WHERE table_schema = ?
	            AND table_name = ?
	          ORDER BY ordinal_position`, dbName, tableName).Scan(&raw)

	var cols []Column
	for _, r := range raw {
		goType := mysqlToGoType(r.DataType, r.ColumnType, r.IsNullable == "YES")
		tag := buildGormTag(r.ColumnName, r.ColumnKey)
		cols = append(cols, Column{
			GoName: toGoName(r.ColumnName),
			GoType: goType,
			Tag:    tag,
			GoTag:  r.ColumnName,
		})
	}
	return cols
}

// ---------- 类型映射 ----------
func mysqlToGoType(dataType, columnType string, nullable bool) string {
	var goType string
	switch dataType {
	case "tinyint", "smallint", "mediumint", "int", "integer":
		goType = "int32"
		if strings.Contains(columnType, "tinyint(1)") {
			goType = "bool" // 常见布尔
		}
	case "bigint":
		goType = "int64"
	case "float":
		goType = "float32"
	case "double", "decimal", "numeric":
		goType = "float64"
	case "char", "varchar", "tinytext", "text", "mediumtext", "longtext", "enum", "set":
		goType = "string"
	case "date", "datetime", "timestamp", "time":
		goType = "time.Time"
	case "json":
		goType = "datatypes.JSON" // gorm 官方 datatypes
	case "binary", "varbinary", "blob", "mediumblob", "longblob":
		goType = "[]byte"
	default:
		goType = "interface{}" // 兜底
	}
	if nullable && !strings.HasPrefix(goType, "[]") && goType != "interface{}" {
		goType = "*" + goType
	}
	return goType
}

// ---------- 构造 gorm tag ----------
func buildGormTag(colName, colKey string) string {
	tags := []string{fmt.Sprintf(`column:%s`, colName)}
	if colKey == "PRI" {
		tags = append(tags, "primaryKey")
	}
	return strings.Join(tags, ";")
}

// ---------- 下划线转驼峰 ----------
func toGoName(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		parts[i] = strings.Title(p)
	}
	return strings.Join(parts, "")
}
func main() {
	flag.Parse()
	// 如果存在*outPath  啥也不干
	if _, err := os.Stat(*outPath); err == nil {
		log.Fatalf("目录 %s 已存在，请勿重复生成", *outPath)
	}

	// 1. 解析模板 FS
	tplFS, isUser := templateFS()
	modelTpl := parseTemplate(tplFS, "model.tpl")
	daoTpl := parseTemplate(tplFS, "dao.tpl")
	apiTpl := parseTemplate(tplFS, "api.tpl")
	mainTpl := parseTemplate(tplFS, "main.tpl")
	docsTpl := parseTemplate(tplFS, "docs.tpl")

	// 2. 连接数据库、解析表结构（与旧代码完全一致，省略）
	_, tables := loadTables()

	// 3. 创建目录
	for _, dir := range []string{"model", "dao", "api", "docs"} {
		_ = os.MkdirAll(filepath.Join(*outPath, dir), 0755)
	}

	// 4. 生成代码
	for _, t := range tables {
		// model
		writeTemplate(modelTpl, t, filepath.Join(*outPath, "model", t.TableName+".go"))
		// dao
		writeTemplate(daoTpl, map[string]interface{}{"Table": t, "Mod": *module},
			filepath.Join(*outPath, "dao", t.TableName+".go"))
		// api
		writeTemplate(apiTpl, map[string]interface{}{"Table": t, "Mod": *module},
			filepath.Join(*outPath, "api", t.TableName+".go"))
	}

	// 5. 生成swagger docs.go
	writeTemplate(docsTpl, nil, filepath.Join(*outPath, "docs", "docs.go"))

	// 6. 生成 main.go
	writeTemplate(mainTpl, map[string]interface{}{
		"Dsn":    *mysqlDsn,
		"Mod":    *module,
		"Tables": tables,
	}, filepath.Join(*outPath, "main.go"))

	// 7. go mod & swag
	execCmd(*outPath, "go", "mod", "init", *module)
	execCmd(*outPath, "go", "mod", "tidy")
	execCmd(*outPath, "swag", "init", "--generalInfo", "main.go")

	fmt.Printf("✅ 生成完成，模板源=%s\n", map[bool]string{true: "本地 tpl/ 目录（用户自定义）", false: "embed 内置模板"}[isUser])
}

// ---------- 模板工具 ----------
func parseTemplate(fsys fs.FS, name string) *template.Template {
	content := readTemplate(fsys, name)
	return template.Must(template.New(name).
		Funcs(template.FuncMap{"lower": strings.ToLower}).
		Parse(content))
}

func writeTemplate(tpl *template.Template, data any, outFile string) {
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		log.Fatalf("execute %s err:%v", outFile, err)
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("fmt warn:%v", err)
		src = buf.Bytes()
	}
	if err := os.WriteFile(outFile, src, 0644); err != nil {
		log.Fatalf("write %s err:%v", outFile, err)
	}
}

func execCmd(dir string, args ...string) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		log.Printf("cmd %+v err:%v", args, err)
	}
}
