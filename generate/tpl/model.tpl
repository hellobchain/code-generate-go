package model

import (
	"{{.Mod}}/constants"
	{{- range .Table.DaoImports }}
	"{{.}}"
	{{- end }}
)

type {{.Table.GoName}} struct {
	{{- range .Table.Columns }}
	{{.GoName}} {{.GoType}} `gorm:"{{.Tag}}" json:"{{.GoTag}}"`    // {{.Comment}}
	{{- end }}
}

func ({{.Table.GoName}}) TableName() string { return constants.Table{{.Table.UpperTableName}} }