// Validate verifies all required fields for {{.AssociatedType}} are set and validates field constraints.
func (p *{{.AssociatedType}}) Validate() error {
	v := new(Validator)
{{- range .Fields}}
{{- if .Required}}
{{- if eq .Type "string"}}
	v.HasRequiredStr(string(p.{{.Name}}), "{{.JSONName}}")
{{- else if eq .Type "int"}}
	v.HasRequiredNum(p.{{.Name}}, "{{.JSONName}}")
{{- else}}
	v.HasRequiredObj(p.{{.Name}}, "{{.JSONName}}")
{{- end}}
{{- end}}
{{- if .Pattern}}
	v.MatchesPattern(string(p.{{.Name}}), `{{.Pattern}}`, "{{.JSONName}}")
{{- end}}
{{- if .Format}}
	v.ValidFormat(string(p.{{.Name}}), "{{.Format}}", "{{.JSONName}}")
{{- end}}
{{- if .EnumType}}
	ValidEnum(v, p.{{.Name}}, {{.CollectionName}}, "{{.JSONName}}")
{{- end}}
{{- if .IsNested}}
{{- if .IsSlice}}
	for i, item := range p.{{.Name}} {
{{- if .IsPointer}}
		if item != nil {
			if err := item.Validate(); err != nil {
				v.AddError(fmt.Errorf("{{.JSONName}}[%d]: %w", i, err))
			}
		}
{{- else}}
		if err := item.Validate(); err != nil {
			v.AddError(fmt.Errorf("{{.JSONName}}[%d]: %w", i, err))
		}
{{- end}}
	}
{{- else if .IsPointer}}
	if p.{{.Name}} != nil {
		if err := p.{{.Name}}.Validate(); err != nil {
			v.AddError(fmt.Errorf("{{.JSONName}}: %w", err))
		}
	}
{{- else}}
	if err := p.{{.Name}}.Validate(); err != nil {
		v.AddError(fmt.Errorf("{{.JSONName}}: %w", err))
	}
{{- end}}
{{- end}}
{{- end}}
	if !v.IsValid() {
		return fmt.Errorf("validation error:\n%v", v.Error())
	}
	return nil
}
