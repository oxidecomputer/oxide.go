{{- range .Variants}}
{{- if .Pattern}}
var {{.TypeName | lower}}Pattern = regexp.MustCompile(`{{.Pattern}}`)
{{- end}}
{{- end}}

func (v *{{.TypeName}}) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	{{- range .Variants}}
	if detect{{.TypeName}}(s) {
		val := {{.TypeName}}(s)
		v.{{$.ValueFieldName}} = &val
		return nil
	}
	{{- end}}
	return fmt.Errorf("no variant matched for {{.TypeName}}: %q", s)
}

func (v {{.TypeName}}) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.{{.ValueFieldName}})
}

{{- range .Variants}}

func detect{{.TypeName}}(s string) bool {
	{{- if .Format}}
	return {{.Format | formatDetectorFunc}}(s)
	{{- else if .Pattern}}
	return {{.TypeName | lower}}Pattern.MatchString(s)
	{{- else}}
	return false
	{{- end}}
}

func ({{.TypeName}}) {{$.MarkerMethod}}() {}
{{- end}}
