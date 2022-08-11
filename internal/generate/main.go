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
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %v", err)

	}
	p := filepath.Join(filepath.Dir(wd), "../VERSION_OMICRON.txt")
	omicronVersion, err := ioutil.ReadFile(p)
	if err != nil {
		return fmt.Errorf("error retrieving Omicron version: %v", err)
	}
	ov := string(omicronVersion)

	// TODO: actually host the spec here.
	// uri := "https://api.oxide.computer"
	uri := fmt.Sprintf("https://raw.githubusercontent.com/oxidecomputer/omicron/%s/openapi/nexus.json", ov)
	u, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("error parsing url %q: %v", uri, err)
	}

	// Load the open API spec from the URI.
	doc, err := openapi3.NewLoader().LoadFromURI(u)
	if err != nil {
		return fmt.Errorf("error loading openAPI spec from %q: %v", uri, err)
	}

	// Generate the client.go file.
	if err := generateClient(doc); err != nil {
		return err
	}

	// Generate the types.go file.
	if err := generateTypes(doc); err != nil {
		return err
	}

	// Generate the responses.go file.
	if err := generateResponses(doc); err != nil {
		return err
	}

	// Generate the paths.go file.
	if err := generatePaths(doc); err != nil {
		return err
	}

	return nil
}
