// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func Test_generateTypes(t *testing.T) {
	typesSpec := &openapi3.T{
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{
				"DiskIdentifier": &openapi3.SchemaRef{Value: &openapi3.Schema{
					Description: "Parameters for the [`Disk`](omicron_common::api::external::Disk) to be attached or detached to an instance",
					Type:        &openapi3.Types{"object"},
					Properties: openapi3.Schemas{"name": &openapi3.SchemaRef{
						Value: &openapi3.Schema{},
						Ref:   "#/components/schemas/Name"}},
				}},
				"DiskCreate": &openapi3.SchemaRef{Value: &openapi3.Schema{
					Type: &openapi3.Types{"object"},
					Properties: openapi3.Schemas{"disk_source": &openapi3.SchemaRef{
						Value: &openapi3.Schema{AllOf: openapi3.SchemaRefs{
							&openapi3.SchemaRef{
								Value: &openapi3.Schema{},
								Ref:   "#/components/schemas/DiskSource",
							},
						}},
					}},
				}},
				"DiskSource": &openapi3.SchemaRef{Value: &openapi3.Schema{
					OneOf: openapi3.SchemaRefs{
						&openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Description: "Create a disk from a disk snapshot",
								Type:        &openapi3.Types{"object"},
								Properties: openapi3.Schemas{
									"snapshot_id": &openapi3.SchemaRef{
										Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "uuid"},
									},
									"type": &openapi3.SchemaRef{
										Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []interface{}{"snapshot"}},
									},
								},
							},
						},
						&openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Description: "Create a disk from a project image",
								Type:        &openapi3.Types{"object"},
								Properties: openapi3.Schemas{
									"image_id": &openapi3.SchemaRef{
										Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "uuid"},
									},
									"type": &openapi3.SchemaRef{
										Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []interface{}{"image"}},
									},
								},
							},
						},
					},
				}},
			},
		},
	}

	type args struct {
		file string
		spec *openapi3.T
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name:    "fail on non-existent file",
			args:    args{"sdf/gdsf", typesSpec},
			wantErr: "no such file or directory",
		},
		{
			name: "success",
			args: args{"test_utils/types_output", typesSpec},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := generateTypes(tt.args.file, tt.args.spec); err != nil {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			if err := compareFiles("test_utils/types_output_expected", tt.args.file); err != nil {
				t.Error(err)
			}
		})
	}
}

func Test_createTypeObject(t *testing.T) {
	typesSpec := openapi3.Schema{
		Required: []string{"type"},
		Properties: map[string]*openapi3.SchemaRef{
			"snapshot_id": {
				Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "uuid"},
			},
			"type": {
				Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []interface{}{"snapshot"}},
			},
		}}

	type args struct {
		s        openapi3.Schema
		name     string
		typeName string
	}
	tests := []struct {
		name string
		args args
		want TypeTemplate
	}{
		{
			name: "success",
			args: args{typesSpec, "DiskSource", "DiskSourceSnapshot"},
			want: TypeTemplate{
				Description: "Create a disk from a disk snapshot\n//\n// Required fields:\n// - Type",
				Name:        "DiskSourceSnapshot",
				Type:        "struct", Fields: []TypeField{
					{
						Description:       "",
						Name:              "SnapshotId",
						Type:              "string",
						SerializationInfo: "`json:\"snapshot_id,omitempty\" yaml:\"snapshot_id,omitempty\"`",
					},
					{
						Description:       "",
						Name:              "Type",
						Type:              "DiskSourceType",
						SerializationInfo: "`json:\"type\" yaml:\"type\"`",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createTypeObject(&tt.args.s, tt.args.name, tt.args.typeName, "Create a disk from a disk snapshot")
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_createStringEnum(t *testing.T) {
	typesSpec := &openapi3.Schema{Enum: []interface{}{"admin", "collaborator", "viewer"}}
	enums := map[string][]string{}
	type args struct {
		s           *openapi3.Schema
		stringEnums map[string][]string
		name        string
		typeName    string
	}
	tests := []struct {
		name  string
		args  args
		want  map[string][]string
		want1 []TypeTemplate
		want2 []EnumTemplate
	}{
		{
			name:  "success",
			args:  args{typesSpec, enums, "FleetRole", "FleetRole"},
			want:  map[string][]string{"FleetRole": {"admin", "collaborator", "viewer"}},
			want1: []TypeTemplate{{Description: "// FleetRole is the type definition for a FleetRole.", Name: "FleetRole", Type: "string"}},
			want2: []EnumTemplate{
				{Description: "// FleetRoleAdmin represents the FleetRole `\"admin\"`.", Name: "FleetRoleAdmin", ValueType: "const", Value: "FleetRole = \"admin\""},
				{Description: "// FleetRoleCollaborator represents the FleetRole `\"collaborator\"`.", Name: "FleetRoleCollaborator", ValueType: "const", Value: "FleetRole = \"collaborator\""},
				{Description: "// FleetRoleViewer represents the FleetRole `\"viewer\"`.", Name: "FleetRoleViewer", ValueType: "const", Value: "FleetRole = \"viewer\""},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, got1, got2 := createStringEnum(tc.args.s, tc.args.stringEnums, tc.args.name, tc.args.typeName)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.want1, got1)
			assert.Equal(t, tc.want2, got2)
		})
	}
}

func Test_createOneOf(t *testing.T) {
	typeSpec := &openapi3.Schema{
		Description: "The source of the underlying image.",
		OneOf: openapi3.SchemaRefs{
			&openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"object"},
					Properties: map[string]*openapi3.SchemaRef{
						"type": {
							Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []interface{}{"url"}},
						},
						"url": {
							Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
						},
					},
					Required: []string{"type", "url"},
				},
			},
			&openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"object"},
					Properties: map[string]*openapi3.SchemaRef{
						"id": {
							Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "uuid"},
						},
						"type": {
							Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []interface{}{"snapshot"}},
						},
					},
					Required: []string{"id", "type"},
				},
			},
		},
	}

	type args struct {
		s        *openapi3.Schema
		name     string
		typeName string
	}
	tests := []struct {
		name  string
		args  args
		want  []TypeTemplate
		want1 []EnumTemplate
	}{
		{
			name: "success",
			args: args{typeSpec, "ImageSource", "ImageSource"},
			want: []TypeTemplate{
				{
					Description: "// ImageSourceType is the type definition for a ImageSourceType.", Name: "ImageSourceType", Type: "string",
				},
				{
					Description: "// ImageSourceUrl is the type definition for a ImageSourceUrl.\n//\n// Required fields:\n// - Type\n// - Url",
					Name:        "ImageSourceUrl",
					Type:        "struct",
					Fields: []TypeField{
						{
							Description: "", Name: "Type", Type: "ImageSourceType", SerializationInfo: "`json:\"type\" yaml:\"type\"`",
						},
						{
							Description: "", Name: "Url", Type: "string", SerializationInfo: "`json:\"url\" yaml:\"url\"`",
						},
					},
				},
				{
					Description: "// ImageSourceSnapshot is the type definition for a ImageSourceSnapshot.\n//\n// Required fields:\n// - Id\n// - Type",
					Name:        "ImageSourceSnapshot",
					Type:        "struct",
					Fields: []TypeField{
						{
							Description: "", Name: "Id", Type: "string", SerializationInfo: "`json:\"id\" yaml:\"id\"`",
						},
						{
							Description: "", Name: "Type", Type: "ImageSourceType", SerializationInfo: "`json:\"type\" yaml:\"type\"`",
						},
					},
				},
				{
					Description: "// ImageSource is the source of the underlying image.", Name: "ImageSource", Type: "struct", Fields: []TypeField{
						{
							Description: "// Type is the type definition for a Type.", Name: "Type", Type: "ImageSourceType", SerializationInfo: "`json:\"type,omitempty\" yaml:\"type,omitempty\"`",
						},
						{
							Description: "// Url is the type definition for a Url.", Name: "Url", Type: "string", SerializationInfo: "`json:\"url,omitempty\" yaml:\"url,omitempty\"`",
						},
						{
							Description: "// Id is the type definition for a Id.", Name: "Id", Type: "string", SerializationInfo: "`json:\"id,omitempty\" yaml:\"id,omitempty\"`",
						},
					},
				},
			},
			want1: []EnumTemplate{
				{Description: "// ImageSourceTypeUrl represents the ImageSourceType `\"url\"`.", Name: "ImageSourceTypeUrl", ValueType: "const", Value: "ImageSourceType = \"url\""},
				{Description: "// ImageSourceTypeSnapshot represents the ImageSourceType `\"snapshot\"`.", Name: "ImageSourceTypeSnapshot", ValueType: "const", Value: "ImageSourceType = \"snapshot\""},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, got1 := createOneOf(tc.args.s, tc.args.name, tc.args.typeName)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.want1, got1)
		})
	}
}

func Test_createAllOf(t *testing.T) {
	typeSpecAllOf := &openapi3.Schema{
		Title: "v4",
		AllOf: openapi3.SchemaRefs{
			&openapi3.SchemaRef{
				Ref:   "#/components/schemas/Ipv4Range",
				Value: &openapi3.Schema{Enum: []interface{}{}},
			},
		},
	}

	enums := map[string][]string{}

	type args struct {
		s           *openapi3.Schema
		stringEnums map[string][]string
		name        string
		typeName    string
	}
	tests := []struct {
		name string
		args args
		want []TypeTemplate
	}{
		{
			name: "success allOf",
			args: args{typeSpecAllOf, enums, "IpRange", "IpRange"},
			want: []TypeTemplate{
				{
					Description: "// IpRange is the type definition for a IpRange.", Name: "IpRange", Type: "interface{}",
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Run(tc.name, func(t *testing.T) {
				got := createAllOf(tc.args.s, tc.args.stringEnums, tc.args.name, tc.args.typeName)
				assert.Equal(t, tc.want, got)
			})
		})
	}
}

func Test_ValidationTemplate_Render(t *testing.T) {
	tests := []struct {
		name     string
		template ValidationTemplate
		want     string
	}{
		{
			name: "with all field types",
			template: ValidationTemplate{
				AssociatedType:  "CreateUserParams",
				RequiredObjects: []string{"Body"},
				RequiredStrings: []string{"Name", "Email"},
				RequiredNums:    []string{"Age"},
			},
			want: `// Validate verifies all required fields for CreateUserParams are set
func (p *CreateUserParams) Validate() error {
	v := new(Validator)
	v.HasRequiredObj(p.Body, "Body")
	v.HasRequiredStr(string(p.Name), "Name")
	v.HasRequiredStr(string(p.Email), "Email")
	v.HasRequiredNum(p.Age, "Age")
	if !v.IsValid() {
		return fmt.Errorf("validation error:\n%v", v.Error())
	}
	return nil
}
`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.template.Render()
			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_EnumTemplate_Render(t *testing.T) {
	tests := []struct {
		name     string
		template EnumTemplate
		want     string
	}{
		{
			name: "const enum",
			template: EnumTemplate{
				Description: "// FleetRoleAdmin represents the FleetRole `\"admin\"`.",
				Name:        "FleetRoleAdmin",
				ValueType:   "const",
				Value:       "FleetRole = \"admin\"",
			},
			want: `// FleetRoleAdmin represents the FleetRole ` + "`" + `"admin"` + "`" + `.
const FleetRoleAdmin FleetRole = "admin"

`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.template.Render()
			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_TypeTemplate_Render(t *testing.T) {
	nameTag := "`json:\"name,omitempty\" yaml:\"name,omitempty\"`"
	streetTag := "`json:\"street\" yaml:\"street\"`"
	cityTag := "`json:\"city\" yaml:\"city\"`"

	tests := []struct {
		name     string
		template TypeTemplate
		want     string
	}{
		{
			name: "primitive type without fields",
			template: TypeTemplate{
				Description: "// FleetRole is the type definition for a FleetRole.",
				Name:        "FleetRole",
				Type:        "string",
			},
			want: `// FleetRole is the type definition for a FleetRole.
type FleetRole string
`,
		},
		{
			name: "struct type with fields",
			template: TypeTemplate{
				Description: "// DiskIdentifier is the identifier for a disk.",
				Name:        "DiskIdentifier",
				Type:        "struct",
				Fields: []TypeField{
					{
						Name:              "Name",
						Type:              "string",
						SerializationInfo: nameTag,
					},
				},
			},
			want: fmt.Sprintf(`// DiskIdentifier is the identifier for a disk.
type DiskIdentifier struct {
	Name string %s
}

`, nameTag),
		},
		{
			name: "struct type with field descriptions",
			template: TypeTemplate{
				Description: "// Address is an address.",
				Name:        "Address",
				Type:        "struct",
				Fields: []TypeField{
					{
						Description:       "// Street is the street name",
						Name:              "Street",
						Type:              "string",
						SerializationInfo: streetTag,
					},
					{
						Description:       "// City is the city name",
						Name:              "City",
						Type:              "string",
						SerializationInfo: cityTag,
					},
				},
			},
			want: fmt.Sprintf(`// Address is an address.
type Address struct {
	// Street is the street name
	Street string %s
	// City is the city name
	City string %s
}

`, streetTag, cityTag),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.template.Render()
			assert.Equal(t, tc.want, got)
		})
	}
}
