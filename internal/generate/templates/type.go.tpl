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
// variant of the {{.OneOfInfo.ValueField}} field based on the {{.OneOfInfo.Discriminator.Key}} discriminator.
func (v *{{.Name}}) UnmarshalJSON(data []byte) error {
	var raw struct {
		{{.OneOfInfo.Discriminator.Field}}  {{.OneOfInfo.Discriminator.Type}}  `json:"{{.OneOfInfo.Discriminator.Key}}"`
		{{.OneOfInfo.ValueField}} json.RawMessage `json:"{{.OneOfInfo.ValueKey}}"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("unmarshaling {{.Name}}: %w\nJSON: %s", err, string(data))
	}
	switch raw.{{.OneOfInfo.Discriminator.Field}} {
{{- range .OneOfInfo.Variants}}
	case {{.EnumValue}}:
		var val {{.ImplType}}
		if err := json.Unmarshal(data, &val); err != nil {
			return fmt.Errorf("unmarshaling {{$.Name}} variant {{.ImplType}}: %w\nJSON: %s", err, string(data))
		}
		v.{{$.OneOfInfo.ValueField}} = val
{{- end}}
	}
	return nil
}

// MarshalJSON implements json.Marshaler for {{.Name}}, setting the {{.OneOfInfo.Discriminator.Key}}
// discriminator based on the {{.OneOfInfo.ValueField}} variant type.
func (v {{.Name}}) MarshalJSON() ([]byte, error) {
	var discriminator {{.OneOfInfo.Discriminator.Type}}
	var innerValue any
	switch val := v.{{.OneOfInfo.ValueField}}.(type) {
{{- range .OneOfInfo.Variants}}
	case {{.ImplType}}:
		discriminator = {{.EnumValue}}
		innerValue = val.{{$.OneOfInfo.ValueField}}
{{- end}}
	}
	return json.Marshal(struct {
		{{.OneOfInfo.Discriminator.Field}} {{.OneOfInfo.Discriminator.Type}} `json:"{{.OneOfInfo.Discriminator.Key}}"`
		{{.OneOfInfo.ValueField}} any `json:"{{.OneOfInfo.ValueKey}},omitzero"`
	}{
		{{.OneOfInfo.Discriminator.Field}}: discriminator,
		{{.OneOfInfo.ValueField}}: innerValue,
	})
}
{{end}}
{{else if eq .Type "struct"}}
type {{.Name}} struct{}
{{- if .OneOfMarker}}
func ({{.Name}}) {{.OneOfMarker}}() {}
{{end}}
{{else}}
type {{.Name}} {{.Type}}
{{- if .OneOfMarker}}
func ({{.Name}}) {{.OneOfMarker}}() {}
{{end}}
{{end -}}
