package e

var msgFlags = map[ErrCode][]string{
	{{- range .Tables }}
	// {{.GoName}}
	ERROR_CREATE_{{.UpperGoName}}_FAIL: {"创建{{.GoName}}失败", "create {{.GoName}} fail"}, // 创建{{.GoName}}失败
	ERROR_DELETE_{{.UpperGoName}}_FAIL: {"删除{{.GoName}}失败", "delete {{.GoName}} fail"},   // 删除{{.GoName}}失败
	ERROR_UPDATE_{{.UpperGoName}}_FAIL : {"获取{{.GoName}}失败", "update {{.GoName}} fail"}, // 更新{{.GoName}}失败
	ERROR_GET_{{.UpperGoName}}_FAIL   : {"获取{{.GoName}}失败", "get {{.GoName}} fail"},// 获取{{.GoName}}失败
	ERROR_LIST_{{.UpperGoName}}_FAIL : {"获取{{.GoName}}列表失败", "list {{.GoName}} fail"}, // 获取{{.GoName}}列表失败
	{{- end }}
}
