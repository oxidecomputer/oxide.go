package main

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func Test_generateResponses(t *testing.T) {
	respDesc := "Error"
	responsesSpec := &openapi3.T{
		Components: openapi3.Components{
			Responses: openapi3.Responses{
				"Error": &openapi3.ResponseRef{Value: &openapi3.Response{
					Description: &respDesc,
					Content: openapi3.Content{
						"application/json": &openapi3.MediaType{
							Schema: &openapi3.SchemaRef{
								Ref:   "#/components/schemas/Error",
								Value: &openapi3.Schema{},
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
			args:    args{"sdf/gdsf", responsesSpec},
			wantErr: "no such file or directory",
		},
		{
			name: "success",
			args: args{"test_utils/responses_output", responsesSpec},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := generateResponses(tt.args.file, tt.args.spec); err != nil {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			if err := compareFiles("test_utils/responses_output_expected", tt.args.file); err != nil {
				t.Error(err)
			}
		})
	}
}
