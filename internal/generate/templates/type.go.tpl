{{splitDocString .Description}}
{{- if eq .Type "interface"}}
type {{.Name}} interface {
	{{.OneOfMarker}}()
}
{{else if .Fields}}
type {{.Name}} {{.Type}} {
{{- range .Fields}}
{{- if .Description}}
	{{.Description}}
{{- end}}
	{{.Name}} {{.GoType}} {{.StructTag}}
{{- end}}
}
{{- if .OneOfMarker}}
func ({{.Name}}) {{.OneOfMarker}}() {}
{{end -}}
{{if .OneOfInfo}}
// UnmarshalJSON implements json.Unmarshaler for {{.Name}}, selecting the correct
// variant of the {{.OneOfInfo.ValueField}} field based on the {{.OneOfInfo.DiscriminatorField}} discriminator.
func (v *{{.Name}}) UnmarshalJSON(data []byte) error {
	var raw struct {
		{{.OneOfInfo.DiscriminatorField}}  {{.OneOfInfo.DiscriminatorType}}  `json:"{{.OneOfInfo.DiscriminatorKey}}"`
		{{.OneOfInfo.ValueField}} json.RawMessage `json:"{{.OneOfInfo.ValueKey}}"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	v.{{.OneOfInfo.DiscriminatorField}} = raw.{{.OneOfInfo.DiscriminatorField}}
	switch raw.{{.OneOfInfo.DiscriminatorField}} {
{{- range .OneOfInfo.Variants}}
	case {{.EnumValue}}:
		var val {{.ImplType}}
		if err := json.Unmarshal(raw.{{$.OneOfInfo.ValueField}}, &val); err != nil {
			return err
		}
		v.{{$.OneOfInfo.ValueField}} = val
{{- end}}
	}
	return nil
}
{{end}}
{{else}}
type {{.Name}} {{.Type}}
{{- if .OneOfMarker}}
func ({{.Name}}) {{.OneOfMarker}}() {}
{{end}}
{{end -}}
