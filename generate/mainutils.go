package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ---------- 表/列元数据结构 ----------
type Column struct {
	GoName string // 结构体字段名
	GoType string // go 类型
	Tag    string // gorm tag
	GoTag  string // go tag
}

type Table struct {
	TableName     string
	GoName        string
	UpperGoName   string
	BaseErrorCode int
	Columns       []Column
	DaoImports    []string // dao 层导入依赖包
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
	baseErrorCode := *baseErrorCode
	for _, tn := range tableNames {
		cols, daoImports := loadColumns(db, dbName, tn)
		goName := toGoName(tn)
		tables = append(tables, Table{
			TableName:     tn,
			GoName:        goName,
			UpperGoName:   strings.ToUpper(tn),
			Columns:       cols,
			BaseErrorCode: baseErrorCode,
			DaoImports:    daoImports,
		})
		baseErrorCode += 100
	}
	return db, tables
}

// ---------- 列级解析 ----------
func loadColumns(db *gorm.DB, dbName, tableName string) ([]Column, []string) {
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
	var daoImports []string
	for _, r := range raw {
		goType := mysqlToGoType(r.DataType, r.ColumnType, r.IsNullable == "YES")
		tag := buildGormTag(r.ColumnName, r.ColumnKey)
		cols = append(cols, Column{
			GoName: toGoName(r.ColumnName),
			GoType: goType,
			Tag:    tag,
			GoTag:  r.ColumnName,
		})
		if strings.Contains(goType, "datatypes.JSON") {
			daoImports = append(daoImports, "gorm.io/datatypes")
		}
		if strings.Contains(goType, "time.Time") {
			daoImports = append(daoImports, "time")
		}
	}
	return cols, daoImports
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
