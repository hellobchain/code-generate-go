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
	GoName  string // 结构体字段名
	GoType  string // go 类型
	Tag     string // gorm tag
	GoTag   string // go tag
	Comment string // 列注释
}

type Table struct {
	TableName      string
	UpperTableName string
	GoName         string
	UpperGoName    string
	BaseErrorCode  int
	Columns        []Column
	DaoImports     []string // dao 层导入依赖包
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
		cols, daoImports := loadColumnsWithIndex(db, dbName, tn)
		goName := toGoName(tn)
		tables = append(tables, Table{
			TableName:      tn,
			UpperTableName: toUpperS(tn),
			GoName:         goName,
			UpperGoName:    strings.ToUpper(tn),
			Columns:        cols,
			BaseErrorCode:  baseErrorCode,
			DaoImports:     daoImports,
		})
		baseErrorCode += 100
	}
	return db, tables
}

// 首字母大写
func toUpperS(tableName string) string {
	// 去掉_ 每个字的首字母大写
	s := strings.Split(tableName, "_")
	for i := range s {
		s[i] = strings.ToUpper(s[i][:1]) + s[i][1:]
	}
	return strings.Join(s, "")
}

func loadColumnsWithIndex(db *gorm.DB, dbName, tableName string) ([]Column, []string) {
	// 1. 列信息
	var cols []struct {
		ColumnName    string `gorm:"column:COLUMN_NAME"`
		DataType      string `gorm:"column:DATA_TYPE"`
		ColumnType    string `gorm:"column:COLUMN_TYPE"`
		IsNullable    string `gorm:"column:IS_NULLABLE"`
		ColumnKey     string `gorm:"column:COLUMN_KEY"`
		ColumnComment string `gorm:"column:COLUMN_COMMENT"`
	}
	db.Raw(`SELECT COLUMN_NAME, DATA_TYPE, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY, COLUMN_COMMENT
	          FROM information_schema.columns
	          WHERE table_schema = ? AND table_name = ?
	          ORDER BY ordinal_position`, dbName, tableName).Scan(&cols)

	// 2. 索引信息
	idxMap := make(map[string]map[string]int) // idxName -> columnName -> seqInIndex
	var idxRows []struct {
		IndexName  string `gorm:"column:INDEX_NAME"`
		NonUnique  int    `gorm:"column:NON_UNIQUE"` // 1=普通 0=唯一
		ColumnName string `gorm:"column:COLUMN_NAME"`
		SeqInIndex int    `gorm:"column:SEQ_IN_INDEX"`
	}
	db.Raw(`SELECT INDEX_NAME, NON_UNIQUE, COLUMN_NAME, SEQ_IN_INDEX
	          FROM information_schema.STATISTICS
	          WHERE table_schema = ? AND table_name = ?
	            AND INDEX_NAME != 'PRIMARY'
	          ORDER BY INDEX_NAME, SEQ_IN_INDEX`, dbName, tableName).Scan(&idxRows)
	for _, r := range idxRows {
		if _, ok := idxMap[r.IndexName]; !ok {
			idxMap[r.IndexName] = make(map[string]int)
		}
		idxMap[r.IndexName][r.ColumnName] = r.SeqInIndex
	}

	// 3. 生成 Column
	var result []Column
	var daoImports []string
	for _, c := range cols {
		tagParts := []string{fmt.Sprintf(`column:%s`, c.ColumnName)}
		if c.ColumnKey == "PRI" {
			tagParts = append(tagParts, "primaryKey")
		}

		// 把涉及本列的所有索引写进 tag
		for idxName, colSeq := range idxMap {
			if _, hit := colSeq[c.ColumnName]; hit {
				if len(colSeq) == 1 {
					// 单列索引
					if colSeq[c.ColumnName] == 1 {
						unique := "index"
						if isUnique := db.Raw(`SELECT NON_UNIQUE FROM information_schema.STATISTICS WHERE table_schema = ? AND table_name = ? AND INDEX_NAME = ? LIMIT 1`, dbName, tableName, idxName).RowsAffected; isUnique == 0 {
							tagParts = append(tagParts, "unique")
						} else {
							tagParts = append(tagParts, fmt.Sprintf(`%s:%s`, unique, idxName))
						}
					}
				} else {
					// 组合索引
					if colSeq[c.ColumnName] == 1 {
						tagParts = append(tagParts, fmt.Sprintf(`index:%s`, idxName))
					} else {
						// 列索引
						tagParts = append(tagParts, fmt.Sprintf(`index:%s`, idxName))
					}
				}
			}
		}
		if c.ColumnComment != "" {
			tagParts = append(tagParts, "comment:'"+c.ColumnComment+"'")
		}
		goType := mysqlToGoType(c.DataType, c.ColumnType, c.IsNullable == "YES")
		result = append(result, Column{
			GoName:  toGoName(c.ColumnName),
			GoType:  goType,
			Tag:     strings.Join(tagParts, ";"),
			Comment: c.ColumnComment,
			GoTag:   c.ColumnName,
		})
		if strings.Contains(goType, "datatypes.JSON") {
			daoImports = append(daoImports, "gorm.io/datatypes")
		}
		if strings.Contains(goType, "time.Time") {
			daoImports = append(daoImports, "time")
		}
	}
	return result, daoImports
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
