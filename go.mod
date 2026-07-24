module github.com/oxidecomputer/oxide.go

go 1.25.0

require (
	github.com/getkin/kin-openapi v0.144.0
	github.com/google/go-cmp v0.7.0
	github.com/iancoleman/strcase v0.3.0
	github.com/pelletier/go-toml v1.9.5
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-openapi/jsonpointer v0.22.5 // indirect
	github.com/go-openapi/swag/jsonname v0.25.5 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/oasdiff/yaml v0.1.1 // indirect
	github.com/oasdiff/yaml3 v0.0.14 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract (
	[v0.0.7-rc.1, v0.0.7-rc.6]
	v0.0.7-rc.1
	[v0.0.1, v0.0.23]
	v0.0.1
)
