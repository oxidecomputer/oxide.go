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
										Value: &openapi3.Schema{
											Type:   &openapi3.Types{"string"},
											Format: "uuid",
										},
									},
									"type": &openapi3.SchemaRef{
										Value: &openapi3.Schema{
											Type: &openapi3.Types{"string"},
											Enum: []interface{}{"snapshot"},
										},
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
										Value: &openapi3.Schema{
											Type:   &openapi3.Types{"string"},
											Format: "uuid",
										},
									},
									"type": &openapi3.SchemaRef{
										Value: &openapi3.Schema{
											Type: &openapi3.Types{"string"},
											Enum: []interface{}{"image"},
										},
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

	t.Run("fallback", func(t *testing.T) {
		f := TypeField{
			Name:                "Foo",
			Schema:              &openapi3.SchemaRef{Value: &openapi3.Schema{Description: ""}},
			FallbackDescription: true,
		}
		assert.Equal(t, "// Foo is the type definition for a Foo.", f.Description())
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
			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
			},
			Required: true,
		}
		assert.Equal(t, "`json:\"id\" yaml:\"id\"`", f.StructTag())
	})

	t.Run("nullable array", func(t *testing.T) {
		f := TypeField{
			Name:       "Items",
			MarshalKey: "items",
			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{Type: &openapi3.Types{"array"}, Nullable: true},
			},
			Required: false,
		}
		assert.Equal(t, "`json:\"items\" yaml:\"items\"`", f.StructTag())
	})

	t.Run("nullable", func(t *testing.T) {
		f := TypeField{
			Name:       "Value",
			MarshalKey: "value",
			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Nullable: true},
			},
			Required: false,
		}
		assert.Equal(t, "`json:\"value,omitempty\" yaml:\"value,omitempty\"`", f.StructTag())
	})

	t.Run("omitdirective", func(t *testing.T) {
		f := TypeField{
			Name:       "Value",
			MarshalKey: "value",
			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
			},
			OmitDirective: "omitzero",
		}
		assert.Equal(t, "`json:\"value,omitzero\" yaml:\"value,omitzero\"`", f.StructTag())
	})

	t.Run("default", func(t *testing.T) {
		f := TypeField{
			Name:       "Count",
			Type:       "int",
			MarshalKey: "count",
			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
			},
			Required: false,
		}
		assert.Equal(t, "`json:\"count,omitempty\" yaml:\"count,omitempty\"`", f.StructTag())
	})
}

func TestTypeField_IsPointer(t *testing.T) {
	tests := []struct {
		name     string
		field    TypeField
		expected bool
	}{
		{
			name:     "nil schema",
			field:    TypeField{Name: "Body", Schema: nil},
			expected: false,
		},
		{
			name: "nullable required",
			field: TypeField{
				Name: "Config",
				Type: "SomeConfig",
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &openapi3.Types{"object"}, Nullable: true},
				},
				Required: true,
			},
			expected: true,
		},
		{
			name: "nullable not required",
			field: TypeField{
				Name: "Config",
				Type: "SomeConfig",
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &openapi3.Types{"object"}, Nullable: true},
				},
				Required: false,
			},
			expected: false,
		},
		{
			name: "not nullable required",
			field: TypeField{
				Name: "Config",
				Type: "SomeConfig",
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &openapi3.Types{"object"}, Nullable: false},
				},
				Required: true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.field.IsPointer())
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

	got := createTypeObject(
		&typesSpec,
		"DiskSource",
		"DiskSourceSnapshot",
		"Create a disk from a disk snapshot",
	)

	want := TypeTemplate{
		Name:        "DiskSourceSnapshot",
		Type:        "struct",
		Description: "Create a disk from a disk snapshot\n//\n// Required fields:\n// - Type",
		Fields: []TypeField{
			{Name: "SnapshotId", Type: "string", MarshalKey: "snapshot_id", Required: false},
			{Name: "Type", Type: "DiskSourceType", MarshalKey: "type", Required: true},
		},
	}

	if diff := cmp.Diff(want, got, cmpIgnoreSchema); diff != "" {
		t.Errorf("createTypeObject() mismatch (-want +got):\n%s", diff)
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
			name: "success",
			args: args{typesSpec, enums, "FleetRole", "FleetRole"},
			want: map[string][]string{"FleetRole": {"admin", "collaborator", "viewer"}},
			want1: []TypeTemplate{
				{
					Description: "// FleetRole is the type definition for a FleetRole.",
					Name:        "FleetRole",
					Type:        "string",
				},
			},
			want2: []EnumTemplate{
				{
					Description: "// FleetRoleAdmin represents the FleetRole `\"admin\"`.",
					Name:        "FleetRoleAdmin",
					ValueType:   "const",
					Value:       "FleetRole = \"admin\"",
				},
				{
					Description: "// FleetRoleCollaborator represents the FleetRole `\"collaborator\"`.",
					Name:        "FleetRoleCollaborator",
					ValueType:   "const",
					Value:       "FleetRole = \"collaborator\"",
				},
				{
					Description: "// FleetRoleViewer represents the FleetRole `\"viewer\"`.",
					Name:        "FleetRoleViewer",
					ValueType:   "const",
					Value:       "FleetRole = \"viewer\"",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := createStringEnum(
				tt.args.s,
				tt.args.stringEnums,
				tt.args.name,
				tt.args.typeName,
			)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
			assert.Equal(t, tt.want2, got2)
		})
	}
}

func Test_createOneOf(t *testing.T) {
	tests := []struct {
		name      string
		schema    *openapi3.Schema
		typeName  string
		wantTypes []TypeTemplate
		wantEnums []EnumTemplate
	}{
		{
			name: "all variants of same type",
			schema: &openapi3.Schema{
				Description: "The source of the underlying image.",
				OneOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Properties: map[string]*openapi3.SchemaRef{
								"type": {
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"string"},
										Enum: []any{"url"},
									},
								},
								"url": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
							},
							Required: []string{"type", "url"},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Properties: map[string]*openapi3.SchemaRef{
								"id": {
									Value: &openapi3.Schema{
										Type:   &openapi3.Types{"string"},
										Format: "uuid",
									},
								},
								"type": {
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"string"},
										Enum: []any{"snapshot"},
									},
								},
							},
							Required: []string{"id", "type"},
						},
					},
				},
			},
			typeName: "ImageSource",
			wantTypes: []TypeTemplate{
				{
					Description: "// ImageSourceType is the type definition for a ImageSourceType.",
					Name:        "ImageSourceType",
					Type:        "string",
				},
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
						{
							Name:                "Type",
							Type:                "ImageSourceType",
							MarshalKey:          "type",
							FallbackDescription: true,
						},
						{Name: "Url", Type: "string", MarshalKey: "url", FallbackDescription: true},
						{Name: "Id", Type: "string", MarshalKey: "id", FallbackDescription: true},
					},
				},
			},
			wantEnums: []EnumTemplate{
				{
					Description: "// ImageSourceTypeUrl represents the ImageSourceType `\"url\"`.",
					Name:        "ImageSourceTypeUrl",
					ValueType:   "const",
					Value:       "ImageSourceType = \"url\"",
				},
				{
					Description: "// ImageSourceTypeSnapshot represents the ImageSourceType `\"snapshot\"`.",
					Name:        "ImageSourceTypeSnapshot",
					ValueType:   "const",
					Value:       "ImageSourceType = \"snapshot\"",
				},
			},
		},
		{
			name: "variants with different value types",
			schema: &openapi3.Schema{
				Description: "A value that can be an int or a string.",
				OneOf: openapi3.SchemaRefs{
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Properties: map[string]*openapi3.SchemaRef{
								"type": {
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"string"},
										Enum: []any{"int"},
									},
								},
								"value": {
									Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
								},
							},
							Required: []string{"type", "value"},
						},
					},
					&openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"object"},
							Properties: map[string]*openapi3.SchemaRef{
								"type": {
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"string"},
										Enum: []any{"string"},
									},
								},
								"value": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
							},
							Required: []string{"type", "value"},
						},
					},
				},
			},
			typeName: "IntOrString",
			wantTypes: []TypeTemplate{
				// Interface for variant types
				{
					Description: "// intOrStringVariant is implemented by IntOrString variants.",
					Name:        "intOrStringVariant",
					Type:        "interface",
					OneOfMarker: "isIntOrStringVariant",
				},
				{
					Description: "// IntOrStringType is the type definition for a IntOrStringType.",
					Name:        "IntOrStringType",
					Type:        "string",
				},
				{
					Description: "// IntOrStringInt is a variant of IntOrString.",
					Name:        "IntOrStringInt",
					Type:        "struct",
					Fields: []TypeField{
						{Name: "Value", Type: "*int", MarshalKey: "value", Required: true},
					},
					OneOfMarker:     "isIntOrStringVariant",
					OneOfMarkerType: "intOrStringVariant",
				},
				{
					Description: "// IntOrStringString is a variant of IntOrString.",
					Name:        "IntOrStringString",
					Type:        "struct",
					Fields: []TypeField{
						{Name: "Value", Type: "string", MarshalKey: "value", Required: true},
					},
					OneOfMarker:     "isIntOrStringVariant",
					OneOfMarkerType: "intOrStringVariant",
				},
				{
					Description: "// IntOrString is a value that can be an int or a string.",
					Name:        "IntOrString",
					Type:        "struct",
					Fields: []TypeField{
						{Name: "Value", Type: "intOrStringVariant", MarshalKey: "value"},
					},
					OneOfDiscriminator:       "type",
					OneOfDiscriminatorMethod: "Type",
					OneOfDiscriminatorType:   "IntOrStringType",
					OneOfValueField:          "value",
					OneOfValueFieldName:      "Value",
					OneOfVariantType:         "intOrStringVariant",
					OneOfVariants: []OneOfVariant{
						{
							DiscriminatorValue:     "int",
							DiscriminatorEnumValue: "Int",
							TypeName:               "IntOrStringInt",
						},
						{
							DiscriminatorValue:     "string",
							DiscriminatorEnumValue: "String",
							TypeName:               "IntOrStringString",
						},
					},
				},
			},
			wantEnums: []EnumTemplate{
				{
					Description: "// IntOrStringTypeInt represents the IntOrStringType `\"int\"`.",
					Name:        "IntOrStringTypeInt",
					ValueType:   "const",
					Value:       "IntOrStringType = \"int\"",
				},
				{
					Description: "// IntOrStringTypeString represents the IntOrStringType `\"string\"`.",
					Name:        "IntOrStringTypeString",
					ValueType:   "const",
					Value:       "IntOrStringType = \"string\"",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotTypes, gotEnums := createOneOf(tc.schema, tc.typeName, tc.typeName)

			if diff := cmp.Diff(tc.wantTypes, gotTypes, cmpIgnoreSchema); diff != "" {
				t.Errorf("types mismatch (-want +got):\n%s", diff)
			}
			assert.Equal(t, tc.wantEnums, gotEnums)
		})
	}

	t.Run("multiple discriminator keys panics", func(t *testing.T) {
		schema := &openapi3.Schema{
			Description: "Schema with multiple discriminator keys.",
			OneOf: openapi3.SchemaRefs{
				&openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"object"},
						Properties: map[string]*openapi3.SchemaRef{
							"type": {
								Value: &openapi3.Schema{
									Type: &openapi3.Types{"string"},
									Enum: []any{"a"},
								},
							},
							"kind": {
								Value: &openapi3.Schema{
									Type: &openapi3.Types{"string"},
									Enum: []any{"x"},
								},
							},
							"value": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
						},
					},
				},
				&openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"object"},
						Properties: map[string]*openapi3.SchemaRef{
							"type": {
								Value: &openapi3.Schema{
									Type: &openapi3.Types{"string"},
									Enum: []any{"b"},
								},
							},
							"kind": {
								Value: &openapi3.Schema{
									Type: &openapi3.Types{"string"},
									Enum: []any{"y"},
								},
							},
							"value": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
						},
					},
				},
			},
		}

		assert.PanicsWithValue(
			t,
			"[ERROR] Found multiple discriminator properties for type MultiDiscriminator: map[kind:{} type:{}]",
			func() { createOneOf(schema, "MultiDiscriminator", "MultiDiscriminator") },
		)
	})
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
