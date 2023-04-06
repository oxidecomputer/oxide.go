// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func Test_generateResponses(t *testing.T) {
	respDesc := "Error"
	responsesSpec := &openapi3.T{
		Components: &openapi3.Components{
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

func Test_populateResponseType(t *testing.T) {
	desc := "Error"
	respType := openapi3.Response{
		Description: &desc,
		Content: openapi3.Content{
			"application/json": &openapi3.MediaType{
				Schema: &openapi3.SchemaRef{
					Ref:   "#/components/schemas/Error",
					Value: &openapi3.Schema{},
				},
			},
		},
	}
	type args struct {
		name string
		r    *openapi3.Response
	}
	tests := []struct {
		name  string
		args  args
		want  []TypeTemplate
		want1 []EnumTemplate
	}{
		{
			name: "success",
			args: args{"Error", &respType},
			want: []TypeTemplate{
				{
					Description: "// ErrorResponse is the response given when error", Name: "ErrorResponse", Type: "Error",
				},
			},
			want1: []EnumTemplate{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := populateResponseType(tt.args.name, tt.args.r)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}
