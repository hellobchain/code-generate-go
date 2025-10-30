package main

import (
	"log"
	"{{.Mod}}/api"
	"{{.Mod}}/dao"
	"{{.Mod}}/model"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/files"
	_ "{{.Mod}}/docs"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// @title     自动生成接口文档
// @version   1.0
// @host      localhost:8080
// @BasePath  /
func main() {
	dsn := "{{.Dsn}}"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil { log.Fatal(err) }
	{{- range .Tables }}
	db.AutoMigrate(&model.{{.GoName}}{})
	{{- end }}

	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiGroup := r.Group("/api")
	{{- range .Tables }}
	{{lower .GoName}}API := api.New{{.GoName}}Api(dao.New{{.GoName}}Dao(db))
	{{lower .GoName}}API.Register(apiGroup)
	{{- end }}

	r.Run(":8080")
}