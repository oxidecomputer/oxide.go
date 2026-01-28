{{splitDocString .Description}}
{{- if eq .Type "interface"}}
type {{.Name}} interface {
	{{.VariantMarker.Method}}()
}

{{else if eq .Type "marker_only"}}
func ({{.Name}}) {{.VariantMarker.Method}}() {}

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
{{- if eq .Variants.UnionType "tagged"}}

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
{{- else if eq .Variants.UnionType "format"}}

func (v *{{.Name}}) UnmarshalJSON(data []byte) error {
	{{- range .Variants.Variants}}
	// Try {{.TypeName}}
	{
		var candidate {{.TypeName}}
		if err := json.Unmarshal(data, &candidate); err == nil {
			if detect{{.TypeName}}Format(&candidate) {
				v.{{$.Variants.ValueFieldName}} = &candidate
				return nil
			}
		}
	}
	{{- end}}
	return fmt.Errorf("no variant matched for {{.Name}}")
}

func (v {{.Name}}) MarshalJSON() ([]byte, error) {
	if v.{{.Variants.ValueFieldName}} == nil {
		return []byte("null"), nil
	}
	return json.Marshal(v.{{.Variants.ValueFieldName}})
}

{{- range .Variants.Variants}}

func detect{{.TypeName}}Format(v *{{.TypeName}}) bool {
	{{- if .FormatFields}}
	{{- range .FormatFields}}
	if !formatDetectors["{{.Format}}"](v.{{.Name}}) {
		return false
	}
	{{- end}}
	{{- else}}
	_ = v // suppress unused warning
	{{- end}}
	return true
}
{{- end}}

{{- range .Variants.Variants}}

// As{{.TypeSuffix}} attempts to convert the {{$.Name}} to a {{.TypeName}}.
// Returns the variant and true if the conversion succeeded, nil and false otherwise.
func (v {{$.Name}}) As{{.TypeSuffix}}() (*{{.TypeName}}, bool) {
	val, ok := v.{{$.Variants.ValueFieldName}}.(*{{.TypeName}})
	return val, ok
}
{{- end}}
{{- else if eq .Variants.UnionType "pattern"}}

var (
	{{- range .Variants.Variants}}
	{{.TypeName | lower}}Pattern = regexp.MustCompile(`{{.Pattern}}`)
	{{- end}}
)

func (v *{{.Name}}) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	{{- range .Variants.Variants}}
	if {{.TypeName | lower}}Pattern.MatchString(s) {
		val := {{.TypeName}}(s)
		v.{{$.Variants.ValueFieldName}} = &val
		return nil
	}
	{{- end}}
	return fmt.Errorf("no pattern matched for {{.Name}}")
}

func (v {{.Name}}) MarshalJSON() ([]byte, error) {
	if v.{{.Variants.ValueFieldName}} == nil {
		return []byte("null"), nil
	}
	return json.Marshal(v.{{.Variants.ValueFieldName}})
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
{{- end}}

{{else if and (eq .Type "struct") .VariantMarker}}
type {{.Name}} struct{}

func ({{.Name}}) {{.VariantMarker.Method}}() {}

{{else}}
type {{.Name}} {{.Type}}
{{end}}
