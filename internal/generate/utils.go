// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
)

func openGeneratedFile(filename string) (*os.File, error) {
	// Get the current working directory.
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current working directory: %v", err)
	}

	p := filepath.Join(cwd, filename)

	// Create the generated files.
	// Open the file for writing.
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("error creating %q: %v", p, err)
	}

	// Add the header to the package.
	fmt.Fprintf(f, "// Code generated by `%s`. DO NOT EDIT.\n\n", filepath.Base(os.Args[0]))
	fmt.Fprintln(f, "// This Source Code Form is subject to the terms of the Mozilla Public")
	fmt.Fprintln(f, "// License, v. 2.0. If a copy of the MPL was not distributed with this")
	fmt.Fprint(f, "// file, You can obtain one at https://mozilla.org/MPL/2.0/.\n\n")
	fmt.Fprintln(f, "package oxide")
	fmt.Fprintln(f, "")

	return f, nil
}

func isLocalEnum(v *openapi3.SchemaRef) bool {
	return v.Ref == "" && v.Value.Type.Is("string") && len(v.Value.Enum) > 0
}

func isLocalObject(v *openapi3.SchemaRef) bool {
	return v.Ref == "" && v.Value.Type.Is("object") && len(v.Value.Properties) > 0
}

func isObjectArray(v *openapi3.SchemaRef) bool {
	if v.Value.AdditionalProperties.Schema != nil {
		return v.Value.AdditionalProperties.Schema.Value.Type.Is("array")
	}

	return false
}

func isNullableArray(v *openapi3.SchemaRef) bool {
	return v.Value.Type.Is("array") && v.Value.Nullable
}

// formatStringType converts a string schema to a valid Go type.
func formatStringType(t *openapi3.Schema) string {
	var format string
	switch t.Format {
	case "date-time":
		format = "*time.Time"
	case "date":
		format = "*time.Time"
	case "time":
		format = "*time.Time"
	default:
		format = "string"
	}

	return format
}

// toLowerFirstLetter returns the given string with the first letter converted to lower case.
func toLowerFirstLetter(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

func trimStringFromSpace(s string) string {
	if idx := strings.Index(s, " "); idx != -1 {
		return s[:idx]
	}
	return s
}

func containsMatchFirstWord(s []string, str string) bool {
	for _, v := range s {
		if trimStringFromSpace(v) == trimStringFromSpace(str) {
			return true
		}
	}

	return false
}

func isPageParam(s string) bool {
	return s == "nextPage" || s == "pageToken" || s == "limit"
}

// convertToValidGoType converts a schema type to a valid Go type.
func convertToValidGoType(property, typeName string, r *openapi3.SchemaRef) string {
	// Use reference as it is the type
	if r.Ref != "" {
		return getReferenceSchema(r)
	}

	if r.Value.AdditionalProperties.Schema != nil {
		if r.Value.AdditionalProperties.Schema.Ref != "" {
			return getReferenceSchema(r.Value.AdditionalProperties.Schema)
		} else if r.Value.AdditionalProperties.Schema.Value.Items.Ref != "" {
			ref := getReferenceSchema(r.Value.AdditionalProperties.Schema.Value.Items)
			if r.Value.AdditionalProperties.Schema.Value.Items.Value.Type.Is("array") {
				return "[]" + ref
			}
			return ref
		}
	}

	// TODO: Handle AllOf
	if r.Value.AllOf != nil {
		if len(r.Value.AllOf) > 1 {
			fmt.Printf("[WARN] TODO: allOf for %q has more than 1 item\n", property)
			return "TODO"
		}

		return convertToValidGoType(property, "", r.Value.AllOf[0])
	}

	var schemaType string

	if r.Value.Type.Is("string") {
		schemaType = formatStringType(r.Value)
	} else if r.Value.Type.Is("integer") {
		// It is necessary to use pointers for integer types as we need
		// to differentiate between an empty value and a 0.
		schemaType = "*int"
	} else if r.Value.Type.Is("number") {
		schemaType = "float64"
	} else if r.Value.Type.Is("boolean") {
		// Using a pointer here as the json encoder takes false as null
		schemaType = "*bool"
	} else if r.Value.Type.Is("array") {
		reference := getReferenceSchema(r.Value.Items)
		if reference != "" {
			return fmt.Sprintf("[]%s", reference)
		}
		// TODO: handle if it is not a reference.
		schemaType = "[]string"
	} else if r.Value.Type.Is("object") {
		// This is a local object, we make sure there are no duplicates
		// by concactenating the type name and the property name.
		schemaType = typeName + strcase.ToCamel(property)
	} else {
		fmt.Printf("[WARN] TODO: handle type %q for %q, marking as interface{} for now\n", r.Value.Type, property)
		schemaType = "interface{}"
	}

	return schemaType
}

func getReferenceSchema(v *openapi3.SchemaRef) string {
	if v.Ref != "" {
		ref := strings.TrimPrefix(v.Ref, "#/components/schemas/")
		if len(v.Value.Enum) > 0 {
			return strcase.ToCamel(ref)
		}

		return strcase.ToCamel(ref)
	}

	return ""
}

func compareFiles(expected, actual string) error {
	f1, err := os.ReadFile(expected)
	if err != nil {
		return err
	}

	f2, err := os.ReadFile(actual)
	if err != nil {
		return err
	}

	if !bytes.Equal(f1, f2) {
		return fmt.Errorf("%v is not equal to %v", expected, actual)
	}
	return nil
}

// This function is mainly used to avoid having parameters named
// "interface" which is a Go type
func verifyNotAGoType(str string) string {
	if str == "interface" {
		return "itf"
	}

	return str
}

func isNumericType(str string) bool {
	numTypes := []string{"int", "int8", "int16", "int32", "int64", "uint", "uint8",
		"uint16", "uint32", "uint64", "uintptr", "float32", "float64"}
	for _, v := range numTypes {
		if str == v {
			return true
		}
	}
	return false
}

func sliceContains[T comparable](s []T, str T) bool {
	for _, a := range s {
		if a == str {
			return true
		}
	}
	return false
}

func allItemsAreSame[T comparable](a []T) bool {
	for _, v := range a {
		if v != a[0] {
			return false
		}
	}
	return true
}
