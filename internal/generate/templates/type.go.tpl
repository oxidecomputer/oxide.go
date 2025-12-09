{{splitDocString .Description}}
{{- if .Fields}}
type {{.Name}} {{.Type}} {
{{- range .Fields}}
{{- if .Description}}
	{{splitDocString .Description}}
{{- end}}
	{{.Name}} {{.Type}} {{.SerializationInfo}}
{{- end}}
}

{{else}}
type {{.Name}} {{.Type}}
{{end -}}
