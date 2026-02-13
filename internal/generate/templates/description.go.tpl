{{define "description"}}{{if .IsExperimental}}// EXPERIMENTAL: This operation is not yet stable and may change or be removed without notice.
//
{{end}}// {{.FunctionName}}{{if .HasSummary}}: {{.Summary}}{{end}}{{if .HasDescription}}
// {{.Description}}{{end}}{{if .IsListAll}}
//
// This method is a wrapper around the `{{.WrappedFunction}}` method.
// This method returns all the pages at once.{{end}}{{if .IsList}}
//
// To iterate over all pages, use the `{{.FunctionName}}AllPages` method, instead.{{end}}
{{end}}