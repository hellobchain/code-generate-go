package e

// 通用错误信息
const (
	SUCCESS                = 200    //常用error
	INVALID_PARAMS ErrCode = 110101 // 参数错误
	ERROR          ErrCode = 110201 // 错误
	UNKNOWN_ERROR  ErrCode = 110202 // 未知错误
)
const (
	{{- range .Tables }}
	// {{.GoName}}
	BASE_{{.UpperGoName}}_ERROR ErrCode = {{.BaseErrorCode}} // {{.GoName}}基本错误
	ERROR_CREATE_{{.UpperGoName}}_FAIL ErrCode = BASE_{{.UpperGoName}}_ERROR + 1 // 创建{{.GoName}}失败
	ERROR_DELETE_{{.UpperGoName}}_FAIL ErrCode = BASE_{{.UpperGoName}}_ERROR + 2// 删除{{.GoName}}失败
	ERROR_UPDATE_{{.UpperGoName}}_FAIL ErrCode = BASE_{{.UpperGoName}}_ERROR + 3 // 更新{{.GoName}}失败
	ERROR_GET_{{.UpperGoName}}_FAIL   ErrCode = BASE_{{.UpperGoName}}_ERROR + 4 // 获取{{.GoName}}失败
	ERROR_LIST_{{.UpperGoName}}_FAIL ErrCode = BASE_{{.UpperGoName}}_ERROR + 5 // 获取{{.GoName}}列表失败
	{{- end }}
)
