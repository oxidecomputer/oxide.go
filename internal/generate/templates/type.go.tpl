{{splitDocString .Description}}
{{- if eq .Type "interface"}}
type {{.Name}} interface {
	{{.VariantMarker}}()
}

{{else if eq .Type "struct"}}
type {{.Name}} struct {
{{- range .Fields}}
{{- if .Description}}
	{{.Description}}
{{- end}}
	{{.Name}} {{.GoType}} {{.StructTag}}
{{- end}}
}
{{- if .VariantMarker}}

func ({{.Name}}) {{.VariantMarker}}() {}
{{- end}}
{{- if .Union}}

{{.Union.RenderMethods .Name}}

{{- range .Union.Variants}}

// As{{.TypeSuffix}} attempts to convert the {{$.Name}} to a {{.TypeName}}.
// Returns the variant and true if the conversion succeeded, nil and false otherwise.
func (v {{$.Name}}) As{{.TypeSuffix}}() (*{{.TypeName}}, bool) {
	val, ok := v.{{$.Union.ValueFieldName}}.(*{{.TypeName}})
	return val, ok
}
{{- end}}
{{- end}}

{{else}}
type {{.Name}} {{.Type}}
{{end}}
