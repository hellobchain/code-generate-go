package gintool

import (
	"net/http"

	"{{.Mod}}/e"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type ApiResponse struct {
	Code   e.ErrCode   `json:"code"`              // 状态码
	Msg    string      `json:"msg,omitempty"`     // 状态短语
	Data   interface{} `json:"data,omitempty"`    // 数据结果集
	ErrMsg string      `json:"err_msg,omitempty"` // 内部错误详情
}

type ApiListResponse struct {
	Code   e.ErrCode   `json:"code"`              // 状态码
	Msg    string      `json:"msg,omitempty"`     // 状态短语
	Rows   interface{} `json:"rows,omitempty"`    // 数据结果集
	ErrMsg string      `json:"err_msg,omitempty"` // 内部错误详情
	Total  int64       `json:"total"`             // 数据总数
}

type CodeMsg struct {
	Code int    `json:"code,omitempty"` // 状态码
	Msg  string `json:"msg,omitempty"`  // 状态短语
}

func NewCodeMsg(code int, msg string) *CodeMsg {
	return &CodeMsg{
		Code: code,
		Msg:  msg,
	}
}

func responseOutput(c *gin.Context, code e.ErrCode, errMsg string, data interface{}) {
	var httpCode = http.StatusOK
	// if code != e.SUCCESS {
	// 	httpCode = http.StatusBadRequest
	// }
	c.JSON(httpCode, ApiResponse{
		Code:   code,
		Msg:    code.Error(),
		Data:   data,
		ErrMsg: errMsg,
	})
}

func responseOutputHttpCode(c *gin.Context, code e.ErrCode, errMsg string, data interface{}) {
	var httpCode = http.StatusOK
	if code != e.SUCCESS {
		httpCode = http.StatusBadRequest
	}
	c.JSON(httpCode, ApiResponse{
		Code:   code,
		Msg:    code.Error(),
		Data:   data,
		ErrMsg: errMsg,
	})
}

func listResponseOutput(c *gin.Context, code e.ErrCode, errMsg string, rows interface{}, total int64) {
	var httpCode = http.StatusOK
	// if code != e.SUCCESS {
	// 	httpCode = http.StatusBadRequest
	// }
	c.JSON(httpCode, ApiListResponse{
		Code:   code,
		Msg:    code.Error(),
		Rows:   rows,
		ErrMsg: errMsg,
		Total:  total,
	})
}
func ResultOk(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "msg": e.GetMsg(e.SUCCESS)})
}

func ResultOkData(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "data": data, "msg": e.GetMsg(e.SUCCESS)})
}

func ResultCode(ctx *gin.Context, code e.ErrCode) {
	var httpCode = http.StatusOK
	ctx.JSON(httpCode, gin.H{"code": code, "msg": e.GetMsg(code)})
}

func ResultPanicCode(ctx *gin.Context, err interface{}) {
	var httpCode = http.StatusOK
	var code e.ErrCode
	switch codeErr := err.(type) {
	case e.ErrCode:
		code = codeErr
	default:
		code = e.ERROR
	}
	ctx.JSON(httpCode, gin.H{"code": code, "msg": e.GetMsg(code)})
}

func ResultPanicMsg(ctx *gin.Context, msg interface{}) {
	var httpCode = http.StatusOK
	var errMsg string = "系统错误,请联系管理员"
	switch codeErr := msg.(type) {
	case string:
		errMsg = codeErr
	}
	ctx.JSON(httpCode, gin.H{"code": e.ERROR, "msg": errMsg})
}

func ResultPanic(ctx *gin.Context, anyInfo interface{}) {
	var httpCode = http.StatusOK
	switch codeErr := anyInfo.(type) {
	case string:
		ctx.JSON(httpCode, gin.H{"code": e.ERROR, "msg": codeErr})
	case e.ErrCode:
		ctx.JSON(httpCode, gin.H{"code": codeErr, "msg": e.GetMsg(codeErr)})
	case CodeMsg:
		ctx.JSON(httpCode, gin.H{"code": codeErr.Code, "msg": codeErr.Msg})
	case *CodeMsg:
		ctx.JSON(httpCode, gin.H{"code": codeErr.Code, "msg": codeErr.Msg})
	case int: // 兼容int类型
		ctx.JSON(httpCode, gin.H{"code": codeErr, "msg": e.GetMsg(e.ErrCode(codeErr))})
	default:
		ctx.JSON(httpCode, gin.H{"code": e.ERROR, "msg": "系统错误,请联系管理员"})
	}
}

func ResultCodeWithDataHttpCode(ctx *gin.Context, err error, data interface{}) {
	var code e.ErrCode
	var errMsg string
	if err == nil {
		code = e.SUCCESS
	} else {
		errMsg = err.Error()
		err1 := errors.Cause(err)
		switch typedErr := err1.(type) { // 将类型断言的结果赋值给变量
		case e.ErrCode:
			code = typedErr
		default:
			code = e.ERROR
		}
	}

	responseOutputHttpCode(ctx, code, errMsg, data)
}

func ResultCodeWithData(ctx *gin.Context, err error, data interface{}) {
	var code e.ErrCode
	var errMsg string
	if err == nil {
		code = e.SUCCESS
	} else {
		errMsg = err.Error()
		err1 := errors.Cause(err)
		switch typedErr := err1.(type) { // 将类型断言的结果赋值给变量
		case e.ErrCode:
			code = typedErr
		default:
			code = e.ERROR
		}
	}

	responseOutput(ctx, code, errMsg, data)
}

func ResultCodeWithListData(ctx *gin.Context, err error, rows interface{}, total int64) {
	var code e.ErrCode
	var errMsg string
	if err == nil {
		code = e.SUCCESS
	} else {
		errMsg = err.Error()
		err1 := errors.Cause(err)
		switch typedErr := err1.(type) { // 将类型断言的结果赋值给变量
		case e.ErrCode:
			code = typedErr
		default:
			code = e.ERROR
		}
	}

	listResponseOutput(ctx, code, errMsg, rows, total)
}
