func (v {{.TypeName}}) {{.DiscriminatorMethod}}() {{.DiscriminatorType}} {
	switch v.{{.ValueFieldName}}.(type) {
	{{- range .Variants}}
	case {{.TypeName}}, *{{.TypeName}}:
		return {{$.DiscriminatorType}}{{.TypeSuffix}}
	{{- end}}
	default:
		return ""
	}
}

func (v *{{.TypeName}}) UnmarshalJSON(data []byte) error {
	type discriminator struct {
		Type string `json:"{{.Discriminator}}"`
	}
	var d discriminator
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}

	var value {{.VariantType}}
	switch d.Type {
	{{- range .Variants}}
	case "{{.DiscriminatorValue}}":
		value = &{{.TypeName}}{}
	{{- end}}
	default:
		return fmt.Errorf("unknown variant %q, expected {{range $i, $v := .Variants}}{{if $i}} or {{end}}'{{.DiscriminatorValue}}'{{end}}", d.Type)
	}
	if err := json.Unmarshal(data, value); err != nil {
		return err
	}
	v.{{.ValueFieldName}} = value
	return nil
}

func (v {{.TypeName}}) MarshalJSON() ([]byte, error) {
	m := make(map[string]any)
	m["{{.Discriminator}}"] = v.{{.DiscriminatorMethod}}()
	if v.{{.ValueFieldName}} != nil {
		valueBytes, err := json.Marshal(v.{{.ValueFieldName}})
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

