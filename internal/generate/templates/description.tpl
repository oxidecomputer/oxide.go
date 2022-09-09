{{define "description"}}// {{.FunctionName}}{{if .HasSummary}}: {{.Summary}}{{end}}{{if .HasDescription}}
// {{.Description}}{{end}}{{if .IsListAll}}
//
// This method is a wrapper around the {{.WrappedFunction}} method.
// This method returns all the pages at once.{{end}}{{if .IsList}}
//
// To iterate over all pages, use the `{{.FunctionName}}AllPages` method, instead.{{end}}{{if .HasParams}}
//
// Parameters
{{range $k, $v := .SignatureParams}}// - `{{$k}}` {{$v.Description}}
{{end}}{{else}}
{{end}}{{end}}