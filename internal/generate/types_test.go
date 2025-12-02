// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
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
										Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []any{"snapshot"}},
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
										Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []any{"image"}},
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
				Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []any{"snapshot"}},
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
				Type:        "struct", Fields: []TypeFields{
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
	typesSpec := &openapi3.Schema{Enum: []any{"admin", "collaborator", "viewer"}}
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := createStringEnum(tt.args.s, tt.args.stringEnums, tt.args.name, tt.args.typeName)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
			assert.Equal(t, tt.want2, got2)
		})
	}
}

func Test_createOneOf(t *testing.T) {
	type args struct {
		s        *openapi3.Schema
		name     string
		typeName string
	}
	tests := []struct {
		name      string
		args      args
		wantTypes []TypeTemplate
		wantEnums []EnumTemplate
	}{
		{
			name: "success: all variants of same type",
			args: args{
				s: &openapi3.Schema{
					Description: "The source of the underlying image.",
					OneOf: openapi3.SchemaRefs{
						&openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"object"},
								Properties: map[string]*openapi3.SchemaRef{
									"type": {
										Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []any{"url"}},
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
										Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []any{"snapshot"}},
									},
								},
								Required: []string{"id", "type"},
							},
						},
					},
				},
				name:     "ImageSource",
				typeName: "ImageSource",
			},
			wantTypes: []TypeTemplate{
				{
					Description: "// ImageSourceType is the type definition for a ImageSourceType.", Name: "ImageSourceType", Type: "string",
				},
				{
					Description: "// ImageSourceUrl is the type definition for a ImageSourceUrl.\n//\n// Required fields:\n// - Type\n// - Url",
					Name:        "ImageSourceUrl",
					Type:        "struct",
					Fields: []TypeFields{
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
					Fields: []TypeFields{
						{
							Description: "", Name: "Id", Type: "string", SerializationInfo: "`json:\"id\" yaml:\"id\"`",
						},
						{
							Description: "", Name: "Type", Type: "ImageSourceType", SerializationInfo: "`json:\"type\" yaml:\"type\"`",
						},
					},
				},
				{
					Description: "// ImageSource is the source of the underlying image.", Name: "ImageSource", Type: "struct", Fields: []TypeFields{
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
			wantEnums: []EnumTemplate{
				{Description: "// ImageSourceTypeUrl represents the ImageSourceType `\"url\"`.", Name: "ImageSourceTypeUrl", ValueType: "const", Value: "ImageSourceType = \"url\""},
				{Description: "// ImageSourceTypeSnapshot represents the ImageSourceType `\"snapshot\"`.", Name: "ImageSourceTypeSnapshot", ValueType: "const", Value: "ImageSourceType = \"snapshot\""},
			},
		},
		{
			name: "success: variants use different types",
			args: args{
				s: &openapi3.Schema{
					Description: "A value that can be a string or an integer.",
					OneOf: openapi3.SchemaRefs{
						&openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"object"},
								Properties: map[string]*openapi3.SchemaRef{
									"type": {
										Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []any{"string"}},
									},
									"value": {
										Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
									},
								},
							},
						},
						&openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{"object"},
								Properties: map[string]*openapi3.SchemaRef{
									"type": {
										Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []any{"integer"}},
									},
									"value": {
										Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
									},
								},
							},
						},
					},
				},
				name:     "FieldValue",
				typeName: "FieldValue",
			},
			wantTypes: []TypeTemplate{
				{
					Description: "// FieldValueType is the type definition for a FieldValueType.",
					Name:        "FieldValueType",
					Type:        "string",
				},
				{
					Description: "// FieldValueString is the type definition for a FieldValueString.",
					Name:        "FieldValueString",
					Type:        "struct",
					Fields: []TypeFields{
						{Name: "Type", Type: "FieldValueType", SerializationInfo: "`json:\"type,omitempty\" yaml:\"type,omitempty\"`"},
						{Name: "Value", Type: "string", SerializationInfo: "`json:\"value,omitempty\" yaml:\"value,omitempty\"`"},
					},
				},
				{
					Description: "// FieldValueInteger is the type definition for a FieldValueInteger.",
					Name:        "FieldValueInteger",
					Type:        "struct",
					Fields: []TypeFields{
						{Name: "Type", Type: "FieldValueType", SerializationInfo: "`json:\"type,omitempty\" yaml:\"type,omitempty\"`"},
						{Name: "Value", Type: "*int", SerializationInfo: "`json:\"value,omitempty\" yaml:\"value,omitempty\"`"},
					},
				},
				{
					Description: "// FieldValue is a value that can be a string or an integer.",
					Name:        "FieldValue",
					Type:        "struct",
					Fields: []TypeFields{
						{Description: "// Type is the type definition for a Type.", Name: "Type", Type: "FieldValueType", SerializationInfo: "`json:\"type,omitempty\" yaml:\"type,omitempty\"`"},
						{Description: "// Value is the type definition for a Value.", Name: "Value", Type: "any", SerializationInfo: "`json:\"value,omitempty\" yaml:\"value,omitempty\"`"},
					},
				},
			},
			wantEnums: []EnumTemplate{
				{Description: "// FieldValueTypeString represents the FieldValueType `\"string\"`.", Name: "FieldValueTypeString", ValueType: "const", Value: "FieldValueType = \"string\""},
				{Description: "// FieldValueTypeInteger represents the FieldValueType `\"integer\"`.", Name: "FieldValueTypeInteger", ValueType: "const", Value: "FieldValueType = \"integer\""},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, got1 := createOneOf(tc.args.s, tc.args.name, tc.args.typeName)
			assert.Equal(t, tc.wantTypes, got)
			assert.Equal(t, tc.wantEnums, got1)
		})
	}

	t.Run("panics on multiple discriminator properties", func(t *testing.T) {
		schema := &openapi3.Schema{
			OneOf: openapi3.SchemaRefs{
				&openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"object"},
						Properties: map[string]*openapi3.SchemaRef{
							"type": {
								Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []any{"foo"}},
							},
							"kind": {
								Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []any{"bar"}},
							},
						},
					},
				},
			},
		}
		assert.PanicsWithValue(t,
			"[ERROR] Found multiple discriminator properties for type BadType: map[kind:{} type:{}]",
			func() { createOneOf(schema, "BadType", "BadType") },
		)
	})
}

func Test_createAllOf(t *testing.T) {
	typeSpecAllOf := &openapi3.Schema{
		Title: "v4",
		AllOf: openapi3.SchemaRefs{
			&openapi3.SchemaRef{
				Ref:   "#/components/schemas/Ipv4Range",
				Value: &openapi3.Schema{Enum: []any{}},
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
					Description: "// IpRange is the type definition for a IpRange.", Name: "IpRange", Type: "any",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				got := createAllOf(tt.args.s, tt.args.stringEnums, tt.args.name, tt.args.typeName)
				assert.Equal(t, tt.want, got)
			})
		})
	}
}
