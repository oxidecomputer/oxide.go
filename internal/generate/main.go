// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"io"
	"net/http"
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

	sdkVersionFile := "../VERSION"
	sdkVersion, err := loadSDKVersionFromFile(sdkVersionFile)
	if err != nil {
		return err
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
	if err := generateVersion(versionFile, spec, sdkVersion); err != nil {
		return err
	}

	return nil
}

func loadSDKVersionFromFile(file string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current working directory: %w", err)
	}

	f := filepath.Join(filepath.Dir(wd), file)
	version, err := os.ReadFile(f)
	if err != nil {
		return "", fmt.Errorf("error retrieving SDK version: %w", err)
	}

	sdkVersion := strings.TrimSpace(string(version))
	if sdkVersion == "" {
		return "", fmt.Errorf("sdk version cannot be empty: %s", f)
	}

	return sdkVersion, nil
}

func loadAPIFromFile(file string) (*openapi3.T, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current working directory: %w", err)
	}

	p := filepath.Join(filepath.Dir(wd), file)
	omicronVersion, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("error retrieving omicron version from %s: %w", p, err)
	}

	ov := strings.TrimSpace(string(omicronVersion))
	if ov == "" {
		return nil, fmt.Errorf("omicron version cannot be empty: %s", p)
	}

	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/oxidecomputer/omicron/%s", ov)
	baseURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing base url %q: %w", rawURL, err)
	}

	latestURL := baseURL.JoinPath("openapi", "nexus", "nexus-latest.json")
	resp, err := http.DefaultClient.Get(latestURL.String())
	if err != nil {
		return nil, fmt.Errorf("error fetching latest openapi file from %q: %w", latestURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("unexpected status code %d fetching %q: %s",
			resp.StatusCode, latestURL, strings.TrimSpace(string(body)))
	}

	var versionedFile strings.Builder
	if _, err := io.Copy(&versionedFile, resp.Body); err != nil {
		return nil, fmt.Errorf("error reading versioned openapi filename from %q: %w", latestURL, err)
	}

	versioned := strings.TrimSpace(versionedFile.String())
	if versioned == "" {
		return nil, fmt.Errorf("versioned filename is empty in %q", latestURL)
	}

	versionedURL := baseURL.JoinPath("openapi", "nexus", versioned)
	doc, err := openapi3.NewLoader().LoadFromURI(versionedURL)
	if err != nil {
		return nil, fmt.Errorf("error loading versioned openapi specification from %q: %w", versionedURL, err)
	}

	return doc, nil
}
