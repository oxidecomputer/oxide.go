{{splitDocString .Description}}
{{- if .Fields}}
type {{.Name}} {{.Type}} {
{{- range .Fields}}
{{- if .Description}}
	{{.Description}}
{{- end}}
	{{.Name}} {{.Type}} {{.StructTag}}
{{- end}}
}

{{else}}
type {{.Name}} {{.Type}}
{{end -}}
