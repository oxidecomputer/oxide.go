{{splitDocString .Description}}
{{- if eq .Type "interface"}}
type {{.Name}} interface {
	{{.MarkerMethod}}()
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
{{- if .ImplementsMarker}}
func ({{.Name}}) {{.ImplementsMarker}}() {}
{{end -}}
{{if .UnmarshalInfo}}
// UnmarshalJSON implements json.Unmarshaler for {{.Name}}.
func (v *{{.Name}}) UnmarshalJSON(data []byte) error {
	var raw struct {
		{{.UnmarshalInfo.DiscriminatorField}}  {{.UnmarshalInfo.DiscriminatorType}}  `json:"{{.UnmarshalInfo.DiscriminatorField | toLower}}"`
		{{.UnmarshalInfo.ValueField}} json.RawMessage `json:"{{.UnmarshalInfo.ValueField | toLower}}"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	v.{{.UnmarshalInfo.DiscriminatorField}} = raw.{{.UnmarshalInfo.DiscriminatorField}}
	switch raw.{{.UnmarshalInfo.DiscriminatorField}} {
{{- range .UnmarshalInfo.Variants}}
	case {{.EnumValue}}:
		var val {{.ImplType}}
		if err := json.Unmarshal(raw.{{$.UnmarshalInfo.ValueField}}, &val); err != nil {
			return err
		}
		v.{{$.UnmarshalInfo.ValueField}} = val
{{- end}}
	}
	return nil
}
{{end}}
{{else}}
type {{.Name}} {{.Type}}
{{- if .ImplementsMarker}}
func ({{.Name}}) {{.ImplementsMarker}}() {}
{{end}}
{{end -}}
