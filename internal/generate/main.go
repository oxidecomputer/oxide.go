package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

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
	file := "../VERSION_OMICRON"

	spec, err := loadApiFromFile(file)
	if err != nil {
		return err
	}

	// Generate the types.go file.
	if err := generateTypes(spec); err != nil {
		return err
	}

	// Generate the responses.go file.
	if err := generateResponses(spec); err != nil {
		return err
	}

	// Generate the paths.go file.
	if err := generatePaths(spec); err != nil {
		return err
	}

	return nil
}

func loadApiFromFile(file string) (*openapi3.T, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current working directory: %v", err)

	}
	p := filepath.Join(filepath.Dir(wd), file)
	omicronVersion, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Omicron version: %v", err)
	}
	ov := string(omicronVersion)

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
