package constants

const (
    {{- range .Tables }}
     Table{{.UpperTableName}} = "{{.TableName}}"
    {{- end }}
)