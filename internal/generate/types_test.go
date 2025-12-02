// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

// cmpIgnoreSchema is a go-cmp option that ignores the Schema field when comparing TypeField.
var cmpIgnoreSchema = cmpopts.IgnoreFields(TypeField{}, "Schema")

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

func TestTypeField_Description(t *testing.T) {
	t.Run("nil schema returns empty", func(t *testing.T) {
		f := TypeField{Name: "Foo", Schema: nil}
		assert.Equal(t, "", f.Description())
	})

	t.Run("empty", func(t *testing.T) {
		f := TypeField{
			Name:   "Foo",
			Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{Description: ""}},
		}
		assert.Equal(t, "", f.Description())
	})

	t.Run("not empty", func(t *testing.T) {
		f := TypeField{
			Name:   "Foo",
			Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{Description: "A foo field"}},
		}
		assert.Equal(t, "// Foo is a foo field", f.Description())
	})
}

func TestTypeField_StructTag(t *testing.T) {
	t.Run("nil schema", func(t *testing.T) {
		f := TypeField{Name: "Body", MarshalKey: "body", Schema: nil}
		assert.Equal(t, "`json:\"body,omitempty\" yaml:\"body,omitempty\"`", f.StructTag())
	})

	t.Run("required", func(t *testing.T) {
		f := TypeField{
			Name:       "Id",
			MarshalKey: "id",
			Schema:     &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			Required:   true,
		}
		assert.Equal(t, "`json:\"id\" yaml:\"id\"`", f.StructTag())
	})

	t.Run("nullable array", func(t *testing.T) {
		f := TypeField{
			Name:       "Items",
			MarshalKey: "items",
			Schema:     &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"array"}, Nullable: true}},
			Required:   false,
		}
		assert.Equal(t, "`json:\"items\" yaml:\"items\"`", f.StructTag())
	})

	t.Run("nullable", func(t *testing.T) {
		f := TypeField{
			Name:       "Value",
			MarshalKey: "value",
			Schema:     &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Nullable: true}},
			Required:   false,
		}
		assert.Equal(t, "`json:\"value,omitempty\" yaml:\"value,omitempty\"`", f.StructTag())
	})

	t.Run("omitdirective", func(t *testing.T) {
		f := TypeField{
			Name:          "Value",
			MarshalKey:    "value",
			Schema:        &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
			OmitDirective: "omitzero",
		}
		assert.Equal(t, "`json:\"value,omitzero\" yaml:\"value,omitzero\"`", f.StructTag())
	})

	t.Run("default", func(t *testing.T) {
		f := TypeField{
			Name:       "Count",
			Type:       "int",
			MarshalKey: "count",
			Schema:     &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}},
			Required:   false,
		}
		assert.Equal(t, "`json:\"count,omitempty\" yaml:\"count,omitempty\"`", f.StructTag())
	})
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

	got := createTypeObject(&typesSpec, "DiskSource", "DiskSourceSnapshot", "Create a disk from a disk snapshot")

	assert.Equal(t, "DiskSourceSnapshot", got.Name)
	assert.Equal(t, "struct", got.Type)
	assert.Equal(t, "Create a disk from a disk snapshot\n//\n// Required fields:\n// - Type", got.Description)
	assert.Len(t, got.Fields, 2)

	// Check first field (snapshot_id)
	assert.Equal(t, "SnapshotId", got.Fields[0].Name)
	assert.Equal(t, "string", got.Fields[0].Type)
	assert.Equal(t, "snapshot_id", got.Fields[0].MarshalKey)
	assert.False(t, got.Fields[0].Required)
	assert.Equal(t, "", got.Fields[0].Description())
	assert.Equal(t, "`json:\"snapshot_id,omitempty\" yaml:\"snapshot_id,omitempty\"`", got.Fields[0].StructTag())

	// Check second field (type)
	assert.Equal(t, "Type", got.Fields[1].Name)
	assert.Equal(t, "DiskSourceType", got.Fields[1].Type)
	assert.Equal(t, "type", got.Fields[1].MarshalKey)
	assert.True(t, got.Fields[1].Required)
	assert.Equal(t, "", got.Fields[1].Description())
	assert.Equal(t, "`json:\"type\" yaml:\"type\"`", got.Fields[1].StructTag())
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
	schema := &openapi3.Schema{
		Description: "The source of the underlying image.",
		OneOf: openapi3.SchemaRefs{
			&openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"object"},
					Properties: map[string]*openapi3.SchemaRef{
						"type": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []any{"url"}}},
						"url":  {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
					},
					Required: []string{"type", "url"},
				},
			},
			&openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"object"},
					Properties: map[string]*openapi3.SchemaRef{
						"id":   {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "uuid"}},
						"type": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []any{"snapshot"}}},
					},
					Required: []string{"id", "type"},
				},
			},
		},
	}

	tests := []struct {
		name      string
		schema    *openapi3.Schema
		typeName  string
		wantTypes []TypeTemplate
		wantEnums []EnumTemplate
	}{
		{
			name:     "all variants of same type",
			schema:   schema,
			typeName: "ImageSource",
			wantTypes: []TypeTemplate{
				{Description: "// ImageSourceType is the type definition for a ImageSourceType.", Name: "ImageSourceType", Type: "string"},
				{
					Description: "// ImageSourceUrl is the type definition for a ImageSourceUrl.\n//\n// Required fields:\n// - Type\n// - Url",
					Name:        "ImageSourceUrl",
					Type:        "struct",
					Fields: []TypeField{
						{Name: "Type", Type: "ImageSourceType", MarshalKey: "type", Required: true},
						{Name: "Url", Type: "string", MarshalKey: "url", Required: true},
					},
				},
				{
					Description: "// ImageSourceSnapshot is the type definition for a ImageSourceSnapshot.\n//\n// Required fields:\n// - Id\n// - Type",
					Name:        "ImageSourceSnapshot",
					Type:        "struct",
					Fields: []TypeField{
						{Name: "Id", Type: "string", MarshalKey: "id", Required: true},
						{Name: "Type", Type: "ImageSourceType", MarshalKey: "type", Required: true},
					},
				},
				{
					Description: "// ImageSource is the source of the underlying image.",
					Name:        "ImageSource",
					Type:        "struct",
					Fields: []TypeField{
						{Name: "Type", Type: "ImageSourceType", MarshalKey: "type", FallbackDescription: true},
						{Name: "Url", Type: "string", MarshalKey: "url", FallbackDescription: true},
						{Name: "Id", Type: "string", MarshalKey: "id", FallbackDescription: true},
					},
				},
			},
			wantEnums: []EnumTemplate{
				{Description: "// ImageSourceTypeUrl represents the ImageSourceType `\"url\"`.", Name: "ImageSourceTypeUrl", ValueType: "const", Value: "ImageSourceType = \"url\""},
				{Description: "// ImageSourceTypeSnapshot represents the ImageSourceType `\"snapshot\"`.", Name: "ImageSourceTypeSnapshot", ValueType: "const", Value: "ImageSourceType = \"snapshot\""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTypes, gotEnums := createOneOf(tt.schema, tt.typeName, tt.typeName)

			if diff := cmp.Diff(tt.wantTypes, gotTypes, cmpIgnoreSchema); diff != "" {
				t.Errorf("types mismatch (-want +got):\n%s", diff)
			}
			assert.Equal(t, tt.wantEnums, gotEnums)
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				got := createAllOf(tt.args.s, tt.args.stringEnums, tt.args.name, tt.args.typeName)
				assert.Equal(t, tt.want, got)
			})
		})
	}
}
