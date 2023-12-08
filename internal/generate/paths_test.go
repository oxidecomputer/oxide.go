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

func Test_generatePaths(t *testing.T) {
	file := "./test_utils/paths.json"
	pathsSpec, err := openapi3.NewLoader().LoadFromFile(file)
	if err != nil {
		t.Error(fmt.Errorf("error loading openAPI spec from %q: %v", file, err))
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
			args:    args{"sdf/gdsf", pathsSpec},
			wantErr: "no such file or directory",
		},
		{
			name: "success",
			args: args{"test_utils/paths_output", pathsSpec},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: For now the test is not properly generating the "ListAll" methods
			// This is because there is a separate check to the response type. The way this works
			// should be changed
			if err := generatePaths(tt.args.file, tt.args.spec); err != nil {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			if err := compareFiles("test_utils/paths_output_expected", tt.args.file); err != nil {
				t.Error(err)
			}
		})
	}
}
