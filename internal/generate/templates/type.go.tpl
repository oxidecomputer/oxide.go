{{splitDocString .Description}}
{{- if eq .Type "interface"}}
type {{.Name}} interface {
	{{.VariantMarker.Method}}()
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

{{- if .VariantMarker}}

func ({{.Name}}) {{.VariantMarker.Method}}() {}
{{- end}}
{{- if .Variants}}

func (v {{.Name}}) {{.Variants.DiscriminatorMethod}}() {{.Variants.DiscriminatorType}} {
	switch v.{{.Variants.ValueFieldName}}.(type) {
	{{- range .Variants.Variants}}
	case *{{.TypeName}}:
		return {{$.Variants.DiscriminatorType}}{{.TypeSuffix}}
	{{- end}}
	default:
		return ""
	}
}

func (v *{{.Name}}) UnmarshalJSON(data []byte) error {
	type discriminator struct {
		Type string `json:"{{.Variants.Discriminator}}"`
	}
	var d discriminator
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}

	var value {{.Variants.VariantType}}
	switch d.Type {
	{{- range .Variants.Variants}}
	case "{{.DiscriminatorValue}}":
		value = &{{.TypeName}}{}
	{{- end}}
	default:
		return fmt.Errorf("unknown variant %q, expected {{range $i, $v := .Variants.Variants}}{{if $i}} or {{end}}'{{.DiscriminatorValue}}'{{end}}", d.Type)
	}
	if err := json.Unmarshal(data, value); err != nil {
		return err
	}
	v.{{.Variants.ValueFieldName}} = value
	return nil
}

func (v {{.Name}}) MarshalJSON() ([]byte, error) {
	m := make(map[string]any)
	m["{{.Variants.Discriminator}}"] = v.{{.Variants.DiscriminatorMethod}}()
	if v.{{.Variants.ValueFieldName}} != nil {
		valueBytes, err := json.Marshal(v.{{.Variants.ValueFieldName}})
		if err != nil {
			return nil, err
		}
		var valueMap map[string]any
		if err := json.Unmarshal(valueBytes, &valueMap); err != nil {
			return nil, err
		}
		for k, val := range valueMap {
			m[k] = val
		}
	}
	return json.Marshal(m)
}

{{- range .Variants.Variants}}

// As{{.TypeSuffix}} attempts to convert the {{$.Name}} to a {{.TypeName}}.
// Returns the variant and true if the conversion succeeded, nil and false otherwise.
func (v {{$.Name}}) As{{.TypeSuffix}}() (*{{.TypeName}}, bool) {
	val, ok := v.{{$.Variants.ValueFieldName}}.(*{{.TypeName}})
	return val, ok
}
{{- end}}
{{- end}}

{{else if and (eq .Type "struct") .VariantMarker}}
type {{.Name}} struct{}

func ({{.Name}}) {{.VariantMarker.Method}}() {}

{{else}}
type {{.Name}} {{.Type}}
{{end}}
