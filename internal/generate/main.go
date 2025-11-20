// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

//go:generate go run ./

func main() {
	if err := generateSDK(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func generateSDK() error {
	// By default, load the Omicron OpenAPI spec from upstream using a version
	// hash specified in `../VERSION_OMICRON`. For local testing, optionally
	// specify a path to an OpenAPI spec file in the `OPENAPI_SPEC_PATH`
	// environment variable, and use its contents instead.
	var spec *openapi3.T
	var err error
	specFileOverride := os.Getenv("OPENAPI_SPEC_PATH")
	if specFileOverride != "" {
		spec, err = openapi3.NewLoader().LoadFromFile(specFileOverride)
		if err != nil {
			return err
		}
	} else {
		versionFile := "../VERSION_OMICRON"
		spec, err = loadAPIFromFile(versionFile)
		if err != nil {
			return err
		}
	}

	typesFile := "../../oxide/types.go"
	if err := generateTypes(typesFile, spec); err != nil {
		return err
	}

	responsesFile := "../../oxide/responses.go"
	if err := generateResponses(responsesFile, spec); err != nil {
		return err
	}

	pathsFile := "../../oxide/paths.go"
	if err := generatePaths(pathsFile, spec); err != nil {
		return err
	}

	versionFile := "../../oxide/version.go"
	if err := generateVersion(versionFile, spec); err != nil {
		return err
	}

	return nil
}

func loadAPIFromFile(file string) (*openapi3.T, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current working directory: %v", err)

	}
	p := filepath.Join(filepath.Dir(wd), file)
	omicronVersion, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Omicron version: %v", err)
	}
	ov := strings.TrimSpace(string(omicronVersion))

	// TODO: actually host the spec here.
	// uri := "https://api.oxide.computer"
	uri := fmt.Sprintf("https://raw.githubusercontent.com/oxidecomputer/omicron/%s/openapi/nexus.json", ov)
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("error parsing url %q: %v", uri, err)
	}

	// Load the open API spec from the URI.
	doc, err := openapi3.NewLoader().LoadFromURI(u)
	if err != nil {
		return nil, fmt.Errorf("error loading openAPI spec from %q: %v", uri, err)
	}

	return doc, nil
}
