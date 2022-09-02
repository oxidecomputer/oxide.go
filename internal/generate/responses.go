package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// TODO: The code generated by this function seems to not be used anywhere. Double check
// Generate the responses.go file.
func generateResponses(file string, spec *openapi3.T) error {
	f, err := openGeneratedFile(file)
	if err != nil {
		return err
	}
	defer f.Close()

	typeCollection := []TypeTemplate{}
	enumCollection := []EnumTemplate{}
	// Iterate over all the responses in the spec and write the types.
	// We want to ensure we keep the order so the diffs don't look like shit.
	keys := make([]string, 0)
	for k := range spec.Components.Responses {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		r := spec.Components.Responses[name]
		if r.Ref != "" {
			fmt.Printf("[WARN] TODO: skipping response for %q, since it is a reference\n", name)
			continue
		}

		tt, et := populateResponseType(name, r.Value)
		typeCollection = append(typeCollection, tt...)
		enumCollection = append(enumCollection, et...)
	}

	writeTypes(f, typeCollection, enumCollection)

	return nil
}

// formatResponseDescription writes the description of the given type.
func formatResponseDescription(name string, r *openapi3.Response) string {
	if r.Description != nil {
		return fmt.Sprintf("// %s is the response given when %s", name, toLowerFirstLetter(
			strings.ReplaceAll(*r.Description, "\n", "\n// ")))
	}

	return fmt.Sprintf("// %s is the type definition for a %s response.\n", name, name)
}

// populateResponseType writes a type definition for the given response.
func populateResponseType(name string, r *openapi3.Response) ([]TypeTemplate, []EnumTemplate) {
	types := []TypeTemplate{}
	enumTypes := []EnumTemplate{}

	for _, v := range r.Content {
		name := fmt.Sprintf("%sResponse", name)

		s := v.Schema
		if s.Ref != "" {
			typeTpl := TypeTemplate{
				Description: formatResponseDescription(name, r),
				Name:        name,
				Type:        getReferenceSchema(s),
			}
			types = append(types, typeTpl)

			continue
		}

		tt, et := populateTypeTemplates(name, s.Value, "")
		types = append(types, tt...)
		enumTypes = append(enumTypes, et...)

	}

	return types, enumTypes
}
