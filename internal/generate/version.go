// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

// generateVersion generates the version.go file with both SDK and API versions.
func generateVersion(file string, spec *openapi3.T) error {
	f, err := openGeneratedFile(file)
	if err != nil {
		return err
	}
	defer f.Close()

	apiVersion := ""
	if spec.Info != nil && spec.Info.Version != "" {
		apiVersion = spec.Info.Version
	}
	if apiVersion == "" {
		return fmt.Errorf("failed generating %s: api version cannnot be empty", file)
	}

	t, err := template.ParseFiles("./templates/version.tpl")
	if err != nil {
		return fmt.Errorf("failed generating %s: %w", file, err)
	}

	data := struct {
		SDKVersion     string
		OpenAPIVersion string
	}{
		SDKVersion:     "v0.8.0",
		OpenAPIVersion: apiVersion,
	}

	if err := t.Execute(f, data); err != nil {
		return fmt.Errorf("failed generating %s: %w", file, err)
	}

	return nil
}
