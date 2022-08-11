package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
)

// TODO: Find a better way to deal with enum types
var collectEnumStringTypes = enumStringTypes()

// TODO: Find a better way to deal with enum types
func enumStringTypes() map[string][]string {
	return map[string][]string{}
}

// Generate the types.go file.
func generateTypes(doc *openapi3.T) error {
	f, err := openGeneratedFile("../../oxide/types.go")
	if err != nil {
		return err
	}
	defer f.Close()

	// Iterate over all the schema components in the spec and write the types.
	// We want to ensure we keep the order so the diffs don't look like shit.
	keys := make([]string, 0)
	for k := range doc.Components.Schemas {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		s := doc.Components.Schemas[name]
		if s.Ref != "" {
			fmt.Printf("[WARN] TODO: skipping type for %q, since it is a reference\n", name)
			continue
		}

		if name == "DatumType" {
			fmt.Printf("[WARN] TODO: skipping type for %q, since it is a duplicate\n", name)
			continue
		}

		writeSchemaType(f, name, s.Value, "")
	}

	// Iterate over all the enum types and add in the slices.
	// We want to ensure we keep the order so the diffs don't look like shit.
	keys = make([]string, 0)
	for k := range collectEnumStringTypes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		enums := collectEnumStringTypes[name]
		// Make the enum a collection of the values.
		// Add a description.
		fmt.Fprintf(f, "// %s is the collection of all %s values.\n", makePlural(name), makeSingular(name))
		fmt.Fprintf(f, "var %s = []%s{\n", makePlural(name), makeSingular(name))
		// We want to keep the values in the same order as the enum.
		sort.Strings(enums)
		for _, enum := range enums {
			// Most likely, the enum values are strings.
			fmt.Fprintf(f, "\t%s,\n", strcase.ToCamel(fmt.Sprintf("%s_%s", makeSingular(name), enum)))
		}
		// Close the enum values.
		fmt.Fprintf(f, "}\n")
	}

	return nil
}

// writeSchemaType writes a type definition for the given schema.
// The additional parameter is only used as a suffix for the type name.
// This is mostly for oneOf types.
func writeSchemaType(f *os.File, name string, s *openapi3.Schema, additionalName string) {
	otype := s.Type
	fmt.Printf("writing type for schema %q -> %s\n", name, otype)

	name = printProperty(name)
	typeName := strings.TrimSpace(fmt.Sprintf("%s%s", name, printProperty(additionalName)))

	if len(s.Enum) == 0 && s.OneOf == nil {
		// Write the type description.
		writeSchemaTypeDescription(typeName, s, f)
	}

	if otype == "string" {
		// If this is an enum, write the enum type.
		if len(s.Enum) > 0 {
			// Make sure we don't redeclare the enum type.
			if _, ok := collectEnumStringTypes[makeSingular(typeName)]; !ok {
				// Write the type description.
				writeSchemaTypeDescription(makeSingular(typeName), s, f)

				// Write the enum type.
				fmt.Fprintf(f, "type %s string\n", makeSingular(typeName))

				collectEnumStringTypes[makeSingular(typeName)] = []string{}
			}

			// Define the enum values.
			fmt.Fprintf(f, "const (\n")
			for _, v := range s.Enum {
				// Most likely, the enum values are strings.
				enum, ok := v.(string)
				if !ok {
					fmt.Printf("[WARN] TODO: enum value is not a string for %q -> %#v\n", name, v)
					continue
				}
				// Write the description of the constant.
				fmt.Fprintf(f, "// %s represents the %s `%q`.\n", strcase.ToCamel(fmt.Sprintf("%s_%s", makeSingular(name), enum)), makeSingular(name), enum)
				fmt.Fprintf(f, "\t%s %s = %q\n", strcase.ToCamel(fmt.Sprintf("%s_%s", makeSingular(name), enum)), makeSingular(name), enum)

				// Add the enum type to the list of enum types.
				collectEnumStringTypes[makeSingular(typeName)] = append(collectEnumStringTypes[makeSingular(typeName)], enum)
			}
			// Close the enum values.
			fmt.Fprintf(f, ")\n")

		} else {
			fmt.Fprintf(f, "type %s string\n", name)
		}
	} else if otype == "integer" {
		fmt.Fprintf(f, "type %s int64\n", name)
	} else if otype == "number" {
		fmt.Fprintf(f, "type %s float64\n", name)
	} else if otype == "boolean" {
		fmt.Fprintf(f, "type %s bool\n", name)
	} else if otype == "array" {
		fmt.Fprintf(f, "type %s []%s\n", name, s.Items.Value.Type)
	} else if otype == "object" {
		recursive := false
		fmt.Fprintf(f, "type %s struct {\n", typeName)
		// We want to ensure we keep the order so the diffs don't look like shit.
		keys := make([]string, 0)
		for k := range s.Properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := s.Properties[k]
			// Check if we need to generate a type for this type.
			typeName := printType(k, v)

			if isLocalEnum(v) {
				recursive = true
				typeName = fmt.Sprintf("%s%s", name, printProperty(k))
			}

			if isLocalObject(v) {
				recursive = true
				fmt.Printf("[WARN] TODO: skipping object for %q -> %#v\n", name, v)
				typeName = fmt.Sprintf("%s%s", name, printProperty(k))
			}

			if v.Value.Description != "" {
				fmt.Fprintf(f, "\t// %s is %s\n", printProperty(k), toLowerFirstLetter(strings.ReplaceAll(v.Value.Description, "\n", "\n// ")))
			}
			fmt.Fprintf(f, "\t%s %s `json:\"%s,omitempty\" yaml:\"%s,omitempty\"`\n", printProperty(k), typeName, k, k)
		}

		fmt.Fprintf(f, "}\n")

		if recursive {
			// Add a newline at the end of the type.
			fmt.Fprintln(f, "")

			// Iterate over the properties and write the types, if we need to.
			for k, v := range s.Properties {
				if isLocalEnum(v) {
					writeSchemaType(f, fmt.Sprintf("%s%s", name, printProperty(k)), v.Value, "")
				}

				if isLocalObject(v) {
					writeSchemaType(f, fmt.Sprintf("%s%s", name, printProperty(k)), v.Value, "")
				}
			}
		}
	} else {
		if s.OneOf != nil {
			// We want to convert these to a different data type to be more idiomatic.
			// But first, we need to make sure we have a type for each one.
			var oneOfTypes []string
			var properties []string
			for _, v := range s.OneOf {
				// We want to iterate over the properties of the embedded object
				// and find the type that is a string.
				var typeName string

				// Iterate over all the schema components in the spec and write the types.
				// We want to ensure we keep the order so the diffs don't look like shit.
				keys := make([]string, 0)
				for k := range v.Value.Properties {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, prop := range keys {
					p := v.Value.Properties[prop]
					// We want to collect all the unique properties to create our global oneOf type.
					propertyName := printType(prop, p)

					propertyString := fmt.Sprintf("\t%s %s `json:\"%s,omitempty\" yaml:\"%s,omitempty\"`\n", printProperty(prop), propertyName, prop, prop)
					if !containsMatchFirstWord(properties, propertyString) {
						properties = append(properties, propertyString)
					}

					if p.Value.Type == "string" {
						if p.Value.Enum != nil {
							// We want to get the enum value.
							// Make sure there is only one.
							if len(p.Value.Enum) != 1 {
								fmt.Printf("[WARN] TODO: oneOf for %q -> %q enum %#v\n", name, prop, p.Value.Enum)
								continue
							}

							typeName = printProperty(p.Value.Enum[0].(string))
						}
					}

					if len(typeName) == 0 && len(keys) == 1 && v.Value.Required != nil && len(v.Value.Required) == 1 {
						typeName = printProperty(v.Value.Required[0])
					}
				}

				// Basically all of these will have one type embedded in them that is a
				// string and the type, since these come from a Rust sum type.
				oneOfType := fmt.Sprintf("%s%s", name, typeName)
				writeSchemaType(f, name, v.Value, typeName)
				// Add it to our array.
				oneOfTypes = append(oneOfTypes, oneOfType)
			}

			// Now let's create the global oneOf type.
			// Write the type description.
			writeSchemaTypeDescription(typeName, s, f)
			fmt.Fprintf(f, "type %s struct {\n", typeName)
			// Iterate over the properties and write the types, if we need to.
			for _, p := range properties {
				fmt.Fprintf(f, p)
			}
			// Close the struct.
			fmt.Fprintf(f, "}\n")

		} else if s.AnyOf != nil {
			fmt.Printf("[WARN] TODO: skipping type for %q, since it is a ANYOF\n", name)
		} else if s.AllOf != nil {
			fmt.Printf("[WARN] TODO: skipping type for %q, since it is a ALLOF\n", name)
		}
	}

	// Add a newline at the end of the type.
	fmt.Fprintln(f, "")
}

// writeSchemaTypeDescription writes the description of the given type.
func writeSchemaTypeDescription(name string, s *openapi3.Schema, f *os.File) {
	if s.Description != "" {
		fmt.Fprintf(f, "// %s is %s\n", name, toLowerFirstLetter(strings.ReplaceAll(s.Description, "\n", "\n// ")))
	} else {
		fmt.Fprintf(f, "// %s is the type definition for a %s.\n", name, name)
	}
}
