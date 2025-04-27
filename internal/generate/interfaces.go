// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
)

// Generate the Interfaces.go file.
func generateInterfaces(file string, methods []methodTemplate) error {
	f, err := openGeneratedFile(file)
	if err != nil {
		return err
	}
	defer f.Close()

	// build the Client interface
	fmt.Fprintln(f, "type Client interface {")
	for _, method := range methods {
		if method.IsListAll {
			fmt.Fprintf(f, "\t%s(ctx context.Context, %s) (%s, error)\n", method.FunctionName, method.ParamsString, method.ResponseType)
		} else if method.ResponseType == "" {
			fmt.Fprintf(f, "\t%s(ctx context.Context, %s) error\n", method.FunctionName, method.ParamsString)
		} else {
			fmt.Fprintf(f, "\t%s(ctx context.Context, %s) (*%s, error)\n", method.FunctionName, method.ParamsString, method.ResponseType)
		}
	}
	fmt.Fprintln(f, "}")
	fmt.Fprintln(f, "")

	return nil
}
