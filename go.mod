module github.com/oxidecomputer/oxide.go

go 1.25.0

require (
	github.com/getkin/kin-openapi v0.153.0
	github.com/google/go-cmp v0.7.0
	github.com/iancoleman/strcase v0.3.0
	github.com/pelletier/go-toml v1.9.5
	github.com/stretchr/testify v1.12.1
)

require (
	github.com/go-openapi/jsonpointer v0.22.5 // indirect
	github.com/go-openapi/swag/jsonname v0.25.5 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	go.yaml.in/yaml/v4 v4.0.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

retract (
	[v0.0.7-rc.1, v0.0.7-rc.6]
	v0.0.7-rc.1
	[v0.0.1, v0.0.23]
	v0.0.1
)
