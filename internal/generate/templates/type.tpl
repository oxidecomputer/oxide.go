{{splitDocString .Description}}
{{if .Fields -}}
type {{.Name}} {{.Type}} {
{{range .Fields -}}
	{{.Render}}
{{- end}}}

{{else -}}
type {{.Name}} {{.Type}}
{{end -}}
