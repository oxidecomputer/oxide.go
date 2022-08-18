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

// Generate the types file.
func generateTypes(file string, spec *openapi3.T) error {
	f, err := openGeneratedFile(file)
	if err != nil {
		return err
	}
	defer f.Close()

	// Iterate over all the schema components in the spec and write the types.
	// We want to ensure we keep the order so the diffs don't look like shit.
	keys := make([]string, 0)
	for k := range spec.Components.Schemas {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		s := spec.Components.Schemas[name]
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
	fmt.Printf("writing type for schema %q -> %s\n", name, s.Type)

	name = printProperty(name)
	typeName := printProperty(name)
	if additionalName != "" {
		typeName = fmt.Sprintf("%s%s", name, printProperty(additionalName))
	}

	switch ot := getObjectType(s); ot {
	case "string":
		writeSchemaTypeDescription(typeName, s, f)
		fmt.Fprintf(f, "type %s string\n", name)
	case "string_enum":
		singularTypename := makeSingular(typeName)
		singularName := makeSingular(name)
		// Make sure we don't redeclare the enum type.
		if _, ok := collectEnumStringTypes[singularTypename]; !ok {
			// Write the type description.
			writeSchemaTypeDescription(singularTypename, s, f)

			// Write the enum type.
			fmt.Fprintf(f, "type %s string\n", singularTypename)

			collectEnumStringTypes[singularTypename] = []string{}
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
			stringType := fmt.Sprintf("%s_%s", singularName, enum)
			fmt.Fprintf(f, "// %s represents the %s `%q`.\n", strcase.ToCamel(stringType), singularName, enum)
			fmt.Fprintf(f, "\t%s %s = %q\n", strcase.ToCamel(stringType), singularName, enum)

			// Add the enum type to the list of enum types.
			collectEnumStringTypes[singularTypename] = append(collectEnumStringTypes[singularTypename], enum)
		}
		// Close the enum values.
		fmt.Fprintf(f, ")\n")
	case "integer":
		writeSchemaTypeDescription(typeName, s, f)
		fmt.Fprintf(f, "type %s int64\n", name)
	case "number":
		writeSchemaTypeDescription(typeName, s, f)
		fmt.Fprintf(f, "type %s float64\n", name)
	case "boolean":
		writeSchemaTypeDescription(typeName, s, f)
		fmt.Fprintf(f, "type %s bool\n", name)
	case "array":
		writeSchemaTypeDescription(typeName, s, f)
		fmt.Fprintf(f, "type %s []%s\n", name, s.Items.Value.Type)
	case "object":
		writeSchemaTypeDescription(typeName, s, f)
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
				typeName = fmt.Sprintf("%s%s", name, printProperty(k))
			}

			if isLocalObject(v) {
				fmt.Printf("[WARN] TODO: skipping object for %q -> %#v\n", name, v)
				typeName = fmt.Sprintf("%s%s", name, printProperty(k))
			}

			if v.Value.Description != "" {
				fmt.Fprintf(f, "\t// %s is %s\n", printProperty(k), toLowerFirstLetter(strings.ReplaceAll(v.Value.Description, "\n", "\n// ")))
			}
			fmt.Fprintf(f, "\t%s %s `json:\"%s,omitempty\" yaml:\"%s,omitempty\"`\n", printProperty(k), typeName, k, k)
		}

		fmt.Fprintf(f, "}\n")

		// Iterate over the properties and write the types, if we need to.
		for k, v := range s.Properties {
			if isLocalEnum(v) {
				writeSchemaType(f, fmt.Sprintf("%s%s", name, printProperty(k)), v.Value, "")
			}

			if isLocalObject(v) {
				writeSchemaType(f, fmt.Sprintf("%s%s", name, printProperty(k)), v.Value, "")
			}
		}
	case "one_of":
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

			writeSchemaType(f, name, v.Value, typeName)
		}

		// Now let's create the global oneOf type.
		// Write the type description.
		writeSchemaTypeDescription(typeName, s, f)
		fmt.Fprintf(f, "type %s struct {\n", typeName)
		// Iterate over the properties and write the types, if we need to.
		for _, p := range properties {
			fmt.Fprint(f, p)
		}
		// Close the struct.
		fmt.Fprintf(f, "}\n")
	case "any_of":
		fmt.Printf("[WARN] TODO: skipping type for %q, since it is a ANYOF\n", name)
	case "all_of":
		fmt.Printf("[WARN] TODO: skipping type for %q, since it is a ALLOF\n", name)
	default:
		fmt.Printf("[WARN] TODO: skipping type for %q, since it is an unknown type\n", name)
	}

	fmt.Fprintln(f, "")
}

func getObjectType(s *openapi3.Schema) string {
	if s.Type == "string" && len(s.Enum) > 0 {
		return "string_enum"
	}

	// TODO: Support enums of other types
	if s.Type != "" {
		return s.Type
	}

	if s.OneOf != nil {
		return "one_of"
	}

	if s.AllOf != nil {
		return "all_of"
	}

	if s.AnyOf != nil {
		return "any_of"
	}

	return ""
}

// writeSchemaTypeDescription writes the description of the given type.
func writeSchemaTypeDescription(name string, s *openapi3.Schema, f *os.File) {
	if s.Description != "" {
		fmt.Fprintf(f, "// %s is %s\n", name, toLowerFirstLetter(strings.ReplaceAll(s.Description, "\n", "\n// ")))
	} else {
		fmt.Fprintf(f, "// %s is the type definition for a %s.\n", name, name)
	}
}
