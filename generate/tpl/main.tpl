package main

import (
	"log"
	"time"
	"{{.Mod}}/api"
	"{{.Mod}}/dao"
	"{{.Mod}}/model"
	"{{.Mod}}/gintool"

	"github.com/gin-gonic/gin"
	_ "{{.Mod}}/docs"
	"github.com/hellobchain/wswlog/wlogging"
	"github.com/jinzhu/gorm"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// 必须要添加，解决找不到mysql驱动问题
	_ "github.com/jinzhu/gorm/dialects/mysql"
)
var logger = wlogging.MustGetFileLoggerWithoutName(nil)
// @title     自动生成接口文档
// @version   1.0
// @host      localhost:8080
// @BasePath  /
func main() {
	wlogging.SetConsole(true)
	dsn := "{{.Dsn}}"
	db, err := gorm.Open("mysql", dsn)
	if err != nil { 
		log.Fatal(err) 
	} else {
		db.DB().SetMaxIdleConns(50)
		db.DB().SetMaxOpenConns(50)
		db.DB().SetConnMaxLifetime(time.Minute)
		db.Set("gorm:association_autoupdate", false).Set("gorm:association_autocreate", false)
		db.SingularTable(true)
		db.LogMode(true)
		db.SetLogger(NewSqlLogger(logger))
	}
	{{- range .Tables }}
	db.AutoMigrate(&model.{{.GoName}}{})
	{{- end }}

	r := gin.New()
	r.Use(gintool.Logger()) // 设置路由日志
	r.Use(gin.Recovery())
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiGroup := r.Group("/api")
	{{- range .Tables }}
	{{lower .GoName}}API := api.New{{.GoName}}Api(dao.New{{.GoName}}Dao(db))
	{{lower .GoName}}API.Register(apiGroup)
	{{- end }}

	r.Run(":8080")
}

type sqlLogger struct {
	logger *wlogging.WswLogger
}

func (l *sqlLogger) Print(v ...interface{}) {
	l.logger.Info(v)
}

func NewSqlLogger(logger *wlogging.WswLogger) *sqlLogger {
	return &sqlLogger{logger: logger}
}