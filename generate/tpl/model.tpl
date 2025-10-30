package model

import (
	"time"
)

type {{.GoName}} struct {
	{{- range .Columns }}
	{{.GoName}} {{.GoType}} `gorm:"{{.Tag}}" json:"{{.GoTag}}"`
	{{- end }}
}

func ({{.GoName}}) TableName() string { return "{{.TableName}}" }