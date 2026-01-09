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
{{- end}}
{{- if .OneOfDiscriminator}}

func (v {{.Name}}) {{.OneOfDiscriminatorMethod}}() {{.OneOfDiscriminatorType}} {
	switch v.{{.OneOfValueFieldName}}.(type) {
	{{- range .OneOfVariants}}
	case *{{.TypeName}}:
		return {{$.OneOfDiscriminatorType}}{{.DiscriminatorEnumValue}}
	{{- end}}
	default:
		return ""
	}
}

func (v *{{.Name}}) UnmarshalJSON(data []byte) error {
	type discriminator struct {
		Type string `json:"{{.OneOfDiscriminator}}"`
	}
	var d discriminator
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}

	var value {{.OneOfVariantType}}
	switch d.Type {
	{{- range .OneOfVariants}}
	case "{{.DiscriminatorValue}}":
		value = &{{.TypeName}}{}
	{{- end}}
	default:
		return fmt.Errorf("unknown variant %q, expected {{range $i, $v := .OneOfVariants}}{{if $i}} or {{end}}'{{.DiscriminatorValue}}'{{end}}", d.Type)
	}
	if err := json.Unmarshal(data, value); err != nil {
		return err
	}
	v.{{.OneOfValueFieldName}} = value
	return nil
}

func (v {{.Name}}) MarshalJSON() ([]byte, error) {
	m := make(map[string]any)
	m["{{.OneOfDiscriminator}}"] = v.{{.OneOfDiscriminatorMethod}}()
	if v.{{.OneOfValueFieldName}} != nil {
		valueBytes, err := json.Marshal(v.{{.OneOfValueFieldName}})
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
{{- end}}

{{else if and (eq .Type "struct") .OneOfMarker}}
type {{.Name}} struct{}

func ({{.Name}}) {{.OneOfMarker}}() {}

{{else}}
type {{.Name}} {{.Type}}
{{end}}
