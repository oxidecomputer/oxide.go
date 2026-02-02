{{- range .Variants}}
{{- range .PatternFields}}
var {{$.TypeName | lower}}{{.Name}}Pattern = regexp.MustCompile(`{{.Pattern}}`)
{{- end}}
{{- end}}

func (v *{{.TypeName}}) UnmarshalJSON(data []byte) error {
	{{- range .Variants}}
	// Try {{.TypeName}}
	{
		var candidate {{.TypeName}}
		if err := json.Unmarshal(data, &candidate); err == nil {
			if detect{{.TypeName}}(&candidate) {
				v.{{$.ValueFieldName}} = &candidate
				return nil
			}
		}
	}
	{{- end}}
	return fmt.Errorf("no variant matched for {{.TypeName}}: %s", string(data))
}

func (v {{.TypeName}}) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.{{.ValueFieldName}})
}

{{- range .Variants}}

func detect{{.TypeName}}(v *{{.TypeName}}) bool {
	{{- range .FormatFields}}
	if !{{.Format | formatDetectorFunc}}(v.{{.Name}}) {
		return false
	}
	{{- end}}
	{{- range .PatternFields}}
	if !{{$.TypeName | lower}}{{.Name}}Pattern.MatchString(v.{{.Name}}) {
		return false
	}
	{{- end}}
	return true
}

func ({{.TypeName}}) {{$.MarkerMethod}}() {}
{{- end}}
