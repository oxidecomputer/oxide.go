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

	typeCollect := []TypeTemplate{}
	enumCollect := []EnumTemplate{}
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

		_, tt, et := writeResponseType(name, r.Value)
		typeCollect = append(typeCollect, tt...)
		enumCollect = append(enumCollect, et...)
	}

	// New code to print to file
	for _, tt := range typeCollect {
		if tt.Name == "" {
			continue
		}

		fmt.Fprintf(f, "%s\n", tt.Description)
		fmt.Fprintf(f, "type %s %s", tt.Name, tt.Type)
		if tt.Fields != nil {
			fmt.Fprint(f, " {\n")
			for _, ft := range tt.Fields {
				if ft.Description != "" {
					// Double check about the "//"
					fmt.Fprintf(f, "\t%s\n", ft.Description)
				}
				fmt.Fprintf(f, "\t%s %s %s\n", ft.Name, ft.Type, ft.SerializationInfo)
			}
			fmt.Fprint(f, "}\n")
		}
		fmt.Fprint(f, "\n")
	}

	for _, et := range enumCollect {
		if et.Name == "" {
			continue
		}

		fmt.Fprintf(f, "%s\n", et.Description)
		fmt.Fprintf(f, "%s %s %s\n", et.ValueType, et.Name, et.Value)
	}

	return nil
}

// writeResponseTypeDescription writes the description of the given type.
func writeResponseTypeDescription(name string, r *openapi3.Response) string {
	if r.Description != nil {
		return fmt.Sprintf("// %s is the response given when %s", name, toLowerFirstLetter(
			strings.ReplaceAll(*r.Description, "\n", "\n// ")))
	}

	return fmt.Sprintf("// %s is the type definition for a %s response.\n", name, name)
}

// writeResponseType writes a type definition for the given response.
func writeResponseType(name string, r *openapi3.Response) (string, []TypeTemplate, []EnumTemplate) {
	var respStr string
	types := []TypeTemplate{}
	enumTypes := []EnumTemplate{}
	// Write the type definition.
	for k, v := range r.Content {
		fmt.Printf("writing type for response %q -> `%s`\n", name, k)

		name := fmt.Sprintf("%sResponse", name)

		// Write the type description.
		respStr = writeResponseTypeDescription(name, r)

		// Print the type definition.
		s := v.Schema
		if s.Ref != "" {
			typeTpl := TypeTemplate{
				Description: writeResponseTypeDescription(name, r),
				Name:        name,
				Type:        getReferenceSchema(s),
			}
			types = append(types, typeTpl)

			// TODO remove once all types are constructed through structs
			respStr = respStr + fmt.Sprintf("type %s %s\n", name, getReferenceSchema(s))
			continue
		}

		// TODO: Ignore the TypeTemplate for now
		// TODO: bubble up printing like types
		resposeType, tt, et := writeSchemaType(name, s.Value, "")
		types = append(types, tt...)
		enumTypes = append(enumTypes, et...)

		respStr = respStr + resposeType
	}

	return respStr, types, enumTypes
}
