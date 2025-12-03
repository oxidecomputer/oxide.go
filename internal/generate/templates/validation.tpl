// Validate verifies all required fields for {{.AssociatedType}} are set
func (p *{{.AssociatedType}}) Validate() error {
	v := new(Validator)
{{- range .RequiredObjects}}
	v.HasRequiredObj(p.{{.}}, "{{.}}")
{{- end}}
{{- range .RequiredStrings}}
	v.HasRequiredStr(string(p.{{.}}), "{{.}}")
{{- end}}
{{- range .RequiredNums}}
	v.HasRequiredNum(p.{{.}}, "{{.}}")
{{- end}}
	if !v.IsValid() {
		return fmt.Errorf("validation error:\n%v", v.Error())}
	return nil
}
