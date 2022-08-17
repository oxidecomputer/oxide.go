package main

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

var successSpec = &openapi3.T{
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
				//	Type: "object",
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
					//		&openapi3.SchemaRef{
					//			Value: &openapi3.Schema{
					//				Description: "Create a disk from a project image",
					//				Type:        "object",
					//				Properties: openapi3.Schemas{"image_id": &openapi3.SchemaRef{
					//					Value: &openapi3.Schema{Type: "string", Enum: []interface{}{"image"}},
					//				}},
					//			},
					//		},
				},
			}},
		},
	},
}

func Test_generateTypes(t *testing.T) {
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
			name:    "success",
			args:    args{"test_generated/types_output.go", successSpec},
			wantErr: "sdfsdf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := generateTypes(tt.args.file, tt.args.spec); err != nil {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			if err := compareFiles("test_utils/types_output_expected.go", tt.args.file); err != nil {
				t.Error(err)
			}
		})
	}
}
