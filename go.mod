module github.com/oxidecomputer/oxide.go

go 1.22.5

require (
	github.com/getkin/kin-openapi v0.129.0
	github.com/iancoleman/strcase v0.3.0
	github.com/pelletier/go-toml v1.9.5
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/oasdiff/yaml v0.0.0-20241210131133-6b86fb107d80 // indirect
	github.com/oasdiff/yaml3 v0.0.0-20241210130736-a94c01f36349 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract (
	[v0.0.7-rc.1, v0.0.7-rc.6]
	v0.0.7-rc.1
	[v0.0.1, v0.0.23]
	v0.0.1
)
