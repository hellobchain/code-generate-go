package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

// flags
var (
	mysqlDsn      = flag.String("dns", "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local", "")
	outPath       = flag.String("out", "../demo", "生成代码目录")
	module        = flag.String("mod", "github.com/demo", "go module 名")
	baseErrorCode = flag.Int("bec", 110300, "")
)

func main() {
	flag.Parse()
	// 如果存在*outPath  啥也不干
	if _, err := os.Stat(*outPath); err == nil {
		log.Fatalf("目录 %s 已存在，请勿重复生成", *outPath)
	}

	// 1. 解析模板 FS
	tplFS, isUser := templateFS()
	// model
	modelTpl := parseTemplate(tplFS, "model.tpl")
	// dao
	daoTpl := parseTemplate(tplFS, "dao.tpl")
	// api
	apiTpl := parseTemplate(tplFS, "api.tpl")
	// main
	mainTpl := parseTemplate(tplFS, "main.tpl")
	// docs
	docsTpl := parseTemplate(tplFS, "docs.tpl")
	// code
	codeTpl := parseTemplate(tplFS, "code.tpl")
	// error
	errorTpl := parseTemplate(tplFS, "error.tpl")
	// msg
	msgTpl := parseTemplate(tplFS, "msg.tpl")
	// result
	resultTpl := parseTemplate(tplFS, "result.tpl")
	// ginlog
	ginlogTpl := parseTemplate(tplFS, "ginlog.tpl")

	// 2. 连接数据库、解析表结构
	_, tables := loadTables()

	// 3. 创建目录
	for _, dir := range []string{"model", "dao", "api", "docs", "e", "gintool"} {
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
	writeTemplate(codeTpl, map[string]interface{}{"Tables": tables}, filepath.Join(*outPath, "e", "code.go"))
	writeTemplate(errorTpl, nil, filepath.Join(*outPath, "e", "error.go"))
	writeTemplate(msgTpl, map[string]interface{}{"Tables": tables}, filepath.Join(*outPath, "e", "msg.go"))
	writeTemplate(resultTpl, map[string]interface{}{"Mod": *module}, filepath.Join(*outPath, "gintool", "result.go"))
	writeTemplate(ginlogTpl, nil, filepath.Join(*outPath, "gintool", "ginlog.go"))

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

	log.Printf("✅ 生成完成，模板源=%s\n", map[bool]string{true: "本地 tpl/ 目录（用户自定义）", false: "embed 内置模板"}[isUser])
}
