package main

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func Test_generateTypes(t *testing.T) {
	typesSpec := &openapi3.T{
		Components: openapi3.Components{
			Schemas: openapi3.Schemas{
				"DiskIdentifier": &openapi3.SchemaRef{Value: &openapi3.Schema{
					Description: "Parameters for the [`Disk`](omicron_common::api::external::Disk) to be attached or detached to an instance",
					Type:        "object",
					Properties: openapi3.Schemas{"name": &openapi3.SchemaRef{
						Value: &openapi3.Schema{},
						Ref:   "#/components/schemas/Name"}},
				}},
				"DiskCreate": &openapi3.SchemaRef{Value: &openapi3.Schema{
					Type: "object",
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
								Type:        "object",
								Properties: openapi3.Schemas{
									"snapshot_id": &openapi3.SchemaRef{
										Value: &openapi3.Schema{Type: "string", Format: "uuid"},
									},
									"type": &openapi3.SchemaRef{
										Value: &openapi3.Schema{Type: "string", Enum: []interface{}{"snapshot"}},
									},
								},
							},
						},
						&openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Description: "Create a disk from a project image",
								Type:        "object",
								Properties: openapi3.Schemas{
									"image_id": &openapi3.SchemaRef{
										Value: &openapi3.Schema{Type: "string", Format: "uuid"},
									},
									"type": &openapi3.SchemaRef{
										Value: &openapi3.Schema{Type: "string", Enum: []interface{}{"image"}},
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
	typesSpec := map[string]*openapi3.SchemaRef{
		"snapshot_id": {
			Value: &openapi3.Schema{Type: "string", Format: "uuid"},
		},
		"type": {
			Value: &openapi3.Schema{Type: "string", Enum: []interface{}{"snapshot"}},
		},
	}

	type args struct {
		s        map[string]*openapi3.SchemaRef
		name     string
		typeName string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want2 TypeTemplate
	}{
		{
			name: "success",
			args: args{typesSpec, "DiskSource", "DiskSourceSnapshot"},
			want: "type DiskSourceSnapshot struct {\n\tSnapshotId string `json:\"snapshot_id,omitempty\" yaml:\"snapshot_id,omitempty\"`\n\tType DiskSourceType `json:\"type,omitempty\" yaml:\"type,omitempty\"`\n}\n",
			want2: TypeTemplate{
				Description: "Create a disk from a disk snapshot",
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
						SerializationInfo: "`json:\"type,omitempty\" yaml:\"type,omitempty\"`",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, typeTpl := createTypeObject(tt.args.s, tt.args.name, tt.args.typeName, "Create a disk from a disk snapshot")
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want2, typeTpl)
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
		want  string
		want1 map[string][]string
	}{
		{
			name:  "success",
			args:  args{typesSpec, enums, "FleetRole", "FleetRole"},
			want:  "// FleetRole is the type definition for a FleetRole.\ntype FleetRole string\nconst (\n// FleetRoleAdmin represents the FleetRole `\"admin\"`.\n\tFleetRoleAdmin FleetRole = \"admin\"\n// FleetRoleCollaborator represents the FleetRole `\"collaborator\"`.\n\tFleetRoleCollaborator FleetRole = \"collaborator\"\n// FleetRoleViewer represents the FleetRole `\"viewer\"`.\n\tFleetRoleViewer FleetRole = \"viewer\"\n)\n",
			want1: map[string][]string{"FleetRole": {"admin", "collaborator", "viewer"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := createStringEnum(tt.args.s, tt.args.stringEnums, tt.args.name, tt.args.typeName)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}

func Test_createOneOf(t *testing.T) {
	typeSpec := &openapi3.Schema{
		Description: "The source of the underlying image.",
		OneOf: openapi3.SchemaRefs{
			&openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: "object",
					Properties: map[string]*openapi3.SchemaRef{
						"type": {
							Value: &openapi3.Schema{Type: "string", Enum: []interface{}{"url"}},
						},
						"url": {
							Value: &openapi3.Schema{Type: "string"},
						},
					},
					Required: []string{"type", "url"},
				},
			},
			&openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: "object",
					Properties: map[string]*openapi3.SchemaRef{
						"id": {
							Value: &openapi3.Schema{Type: "string", Format: "uuid"},
						},
						"type": {
							Value: &openapi3.Schema{Type: "string", Enum: []interface{}{"snapshot"}},
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
		want  string
		want2 []TypeTemplate
	}{
		{
			name: "success",
			args: args{typeSpec, "ImageSource", "ImageSource"},
			want: "// ImageSourceUrl is the type definition for a ImageSourceUrl.\ntype ImageSourceUrl struct {\n\tType ImageSourceType `json:\"type,omitempty\" yaml:\"type,omitempty\"`\n\tUrl string `json:\"url,omitempty\" yaml:\"url,omitempty\"`\n}\n// ImageSourceType is the type definition for a ImageSourceType.\ntype ImageSourceType string\nconst (\n// ImageSourceTypeUrl represents the ImageSourceType `\"url\"`.\n\tImageSourceTypeUrl ImageSourceType = \"url\"\n)\n\n\n// ImageSourceSnapshot is the type definition for a ImageSourceSnapshot.\ntype ImageSourceSnapshot struct {\n\tId string `json:\"id,omitempty\" yaml:\"id,omitempty\"`\n\tType ImageSourceType `json:\"type,omitempty\" yaml:\"type,omitempty\"`\n}\nconst (\n// ImageSourceTypeSnapshot represents the ImageSourceType `\"snapshot\"`.\n\tImageSourceTypeSnapshot ImageSourceType = \"snapshot\"\n)\n\n\n// ImageSource is the source of the underlying image.\ntype ImageSource struct {\n\tType string `json:\"type,omitempty\" yaml:\"type,omitempty\"`\n\tUrl string `json:\"url,omitempty\" yaml:\"url,omitempty\"`\n\tId string `json:\"id,omitempty\" yaml:\"id,omitempty\"`\n}\n",
			want2: []TypeTemplate{
				{
					Description: "", Name: "", Type: ""},
				{
					Description: "// ImageSourceUrl is the type definition for a ImageSourceUrl.", Name: "ImageSourceUrl", Type: "struct", Fields: []TypeFields{
						{
							Description: "", Name: "Type", Type: "ImageSourceType", SerializationInfo: "`json:\"type,omitempty\" yaml:\"type,omitempty\"`"},
						{
							Description: "", Name: "Url", Type: "string", SerializationInfo: "`json:\"url,omitempty\" yaml:\"url,omitempty\"`",
						},
					},
				},
				{
					Description: "", Name: "", Type: "",
				},
				{
					Description: "// ImageSourceSnapshot is the type definition for a ImageSourceSnapshot.", Name: "ImageSourceSnapshot", Type: "struct", Fields: []TypeFields{
						{
							Description: "", Name: "Id", Type: "string", SerializationInfo: "`json:\"id,omitempty\" yaml:\"id,omitempty\"`"},
						{
							Description: "", Name: "Type", Type: "ImageSourceType", SerializationInfo: "`json:\"type,omitempty\" yaml:\"type,omitempty\"`"},
					},
				},
				{
					Description: "// ImageSource is the source of the underlying image.", Name: "ImageSource", Type: "struct", Fields: []TypeFields{
						{
							Description: "", Name: "Type", Type: "string", SerializationInfo: "`json:\"type,omitempty\" yaml:\"type,omitempty\"`"},
						{
							Description: "", Name: "Url", Type: "string", SerializationInfo: "`json:\"url,omitempty\" yaml:\"url,omitempty\"`"},
						{
							Description: "", Name: "Id", Type: "string", SerializationInfo: "`json:\"id,omitempty\" yaml:\"id,omitempty\"`"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, typeTpls := createOneOf(tt.args.s, tt.args.name, tt.args.typeName)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want2, typeTpls)
		})
	}
}
