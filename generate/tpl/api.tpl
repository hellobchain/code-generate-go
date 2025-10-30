package api

import (
	"strconv"
	"{{.Mod}}/e"
	"github.com/gin-gonic/gin"
	"{{.Mod}}/dao"
	"{{.Mod}}/model"
	"{{.Mod}}/gintool"
)

type {{.Table.GoName}}Api struct{ d *dao.{{.Table.GoName}}Dao }

func New{{.Table.GoName}}Api(d *dao.{{.Table.GoName}}Dao) *{{.Table.GoName}}Api { return &{{.Table.GoName}}Api{d: d} }

func (a *{{.Table.GoName}}Api) Register(r *gin.RouterGroup) {
	g := r.Group("/{{.Table.TableName}}")
	g.POST("", a.Create)
	g.GET("/:id", a.Get)
	g.PUT("/:id", a.Update)
	g.DELETE("/:id", a.Delete)
	g.GET("", a.List)
}

// Create
// @Summary  创建{{.Table.GoName}}
// @Tags     {{.Table.TableName}}
// @Accept   json
// @Produce  json
// @Param    body  body     model.{{.Table.GoName}}  true  "实体"
// @Success 200 {object} gintool.ApiResponse{result=model.{{.Table.GoName}}}
// @Failure 400 {object} gintool.ApiResponse
// @Router   /api/{{.Table.TableName}} [post]
func (a *{{.Table.GoName}}Api) Create(c *gin.Context) {
	var m model.{{.Table.GoName}}
	if err := c.ShouldBindJSON(&m); err != nil {
		gintool.ResultCodeWithData(c, e.NewError(e.INVALID_PARAMS, err), nil)
		return
	}
	if err := a.d.Create(&m); err != nil {
		gintool.ResultCodeWithData(c, e.NewError(e.ERROR_CREATE_{{.Table.UpperGoName}}_FAIL, err), nil)
		return
	}
	gintool.ResultCodeWithData(c, nil, m)
}

// Get
// @Summary  根据 ID 获取{{.Table.GoName}}
// @Tags     {{.Table.TableName}}
// @Produce  json
// @Param    id  path  int  true  "主键"
// @Success 200 {object} gintool.ApiResponse{result=model.{{.Table.GoName}}}
// @Failure 400 {object} gintool.ApiResponse
// @Router   /api/{{.Table.TableName}}/{id} [get]
func (a *{{.Table.GoName}}Api) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	m, err := a.d.Get(uint(id))
	if err != nil {
		gintool.ResultCodeWithData(c, e.NewError(e.ERROR_GET_{{.Table.UpperGoName}}_FAIL, err), nil)
		return
	}
	gintool.ResultCodeWithData(c, nil, m)
}

// Update
// @Summary  更新{{.Table.GoName}}
// @Tags     {{.Table.TableName}}
// @Accept   json
// @Produce  json
// @Param    id    path  int              true  "主键"
// @Param    body  body  model.{{.Table.GoName}}  true  "实体"
// @Success 200 {object} gintool.ApiResponse{result=model.{{.Table.GoName}}}
// @Failure 400 {object} gintool.ApiResponse
// @Router   /api/{{.Table.TableName}}/{id} [put]
func (a *{{.Table.GoName}}Api) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var m model.{{.Table.GoName}}
	if err := c.ShouldBindJSON(&m); err != nil {
		gintool.ResultCodeWithData(c, e.NewError(e.INVALID_PARAMS, err), nil)
		return
	}
	m.Id = int64(id)
	if err := a.d.Update(&m); err != nil {
		gintool.ResultCodeWithData(c, e.NewError(e.ERROR_UPDATE_{{.Table.UpperGoName}}_FAIL, err), nil)
		return
	}
	gintool.ResultCodeWithData(c, nil, m)
}

// Delete
// @Summary  删除{{.Table.GoName}}
// @Tags     {{.Table.TableName}}
// @Produce  json
// @Param    id  path  int  true  "主键"
// @Success 200 {object} gintool.ApiResponse{result=model.{{.Table.GoName}}}
// @Failure 400 {object} gintool.ApiResponse
// @Router   /api/{{.Table.TableName}}/{id} [delete]
func (a *{{.Table.GoName}}Api) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := a.d.Delete(uint(id)); err != nil {
		gintool.ResultCodeWithData(c, e.NewError(e.ERROR_DELETE_{{.Table.UpperGoName}}_FAIL, err), nil)
		return
	}
	gintool.ResultCodeWithData(c, nil, nil)
}

// List
// @Summary  获取{{.Table.GoName}}列表
// @Tags     {{.Table.TableName}}
// @Produce  json
// @Success 200 {object} gintool.ApiResponse{result=model.{{.Table.GoName}}}
// @Failure 400 {object} gintool.ApiResponse
// @Router   /api/{{.Table.TableName}} [get]
func (a *{{.Table.GoName}}Api) List(c *gin.Context) {
	list, err := a.d.List()
	if err != nil {
		gintool.ResultCodeWithData(c, e.NewError(e.ERROR_LIST_{{.Table.UpperGoName}}_FAIL, err), nil)
		return
	}
	gintool.ResultCodeWithData(c, nil, list)
}