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

// TODO: use these two structs to build each type

// TypeTemplate holds the information of a type struct
type TypeTemplate struct {
	Description string
	Name        string
	// Type describes the type of the type (e.g. struct, int64, string)
	Type   string
	Fields []TypeFields
}

// TypeFields holds the information for each type field
type TypeFields struct {
	Name              string
	Type              string
	SerializationInfo string
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

	name = strcase.ToCamel(name)
	typeName := strcase.ToCamel(name)
	if additionalName != "" {
		typeName = fmt.Sprintf("%s%s", name, strcase.ToCamel(additionalName))
	}

	switch ot := getObjectType(s); ot {
	case "string":
		writeSchemaTypeDescription(typeName, s, f)
		fmt.Fprintf(f, "type %s string\n", name)
	case "string_enum":
		strEnum, enums := createStringEnum(s, collectEnumStringTypes, name, typeName)
		collectEnumStringTypes = enums
		fmt.Fprint(f, strEnum)
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
		typeObj := createTypeObject(s.Properties, name, typeName)
		fmt.Fprint(f, typeObj)
		// TODO: Handle these differently. The output of these
		// is generally not structs, but constants, variables and string types
		// Iterate over the properties and write the types, if we need to.
		for k, v := range s.Properties {
			if isLocalEnum(v) {
				writeSchemaType(f, fmt.Sprintf("%s%s", name, strcase.ToCamel(k)), v.Value, "")
			}

			if isLocalObject(v) {
				writeSchemaType(f, fmt.Sprintf("%s%s", name, strcase.ToCamel(k)), v.Value, "")
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
				propertyName := convertToValidGoType(prop, p)

				propertyString := fmt.Sprintf("\t%s %s `json:\"%s,omitempty\" yaml:\"%s,omitempty\"`\n", strcase.ToCamel(prop), propertyName, prop, prop)
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

					typeName = strcase.ToCamel(p.Value.Enum[0].(string))
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

// TODO: use the TypeTemplate struct to build these
func createTypeObject(schemas map[string]*openapi3.SchemaRef, name, typeName string) string {
	var typeObj string
	typeObj = fmt.Sprintf("type %s struct {\n", typeName)
	// We want to ensure we keep the order so the diffs don't look like shit.
	keys := make([]string, 0)
	for k := range schemas {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := schemas[k]
		// Check if we need to generate a type for this type.
		typeName := convertToValidGoType(k, v)

		if isLocalEnum(v) {
			typeName = fmt.Sprintf("%s%s", name, strcase.ToCamel(k))
		}

		if isLocalObject(v) {
			fmt.Printf("[WARN] TODO: skipping object for %q -> %#v\n", name, v)
			typeName = fmt.Sprintf("%s%s", name, strcase.ToCamel(k))
		}

		if v.Value.Description != "" {
			typeObj = typeObj + fmt.Sprintf("\t// %s is %s\n", strcase.ToCamel(k), toLowerFirstLetter(strings.ReplaceAll(v.Value.Description, "\n", "\n// ")))
		}
		typeObj = typeObj + fmt.Sprintf("\t%s %s `json:\"%s,omitempty\" yaml:\"%s,omitempty\"`\n", strcase.ToCamel(k), typeName, k, k)

	}

	return typeObj + "}\n"
}

func createStringEnum(s *openapi3.Schema, stringEnums map[string][]string, name, typeName string) (string, map[string][]string) {
	var strEnum string
	singularTypename := makeSingular(typeName)
	singularName := makeSingular(name)
	// Make sure we don't redeclare the enum type.
	if _, ok := stringEnums[singularTypename]; !ok {
		strEnum = schemaTypeDescription(singularTypename, s)

		// Write the enum type.
		strEnum = strEnum + fmt.Sprintf("type %s string\n", singularTypename)

		stringEnums[singularTypename] = []string{}
	}

	// Define the enum values.
	strEnum = strEnum + "const (\n"
	for _, v := range s.Enum {
		// Most likely, the enum values are strings.
		enum, ok := v.(string)
		if !ok {
			fmt.Printf("[WARN] TODO: enum value is not a string for %q -> %#v\n", name, v)
			continue
		}
		// Write the description of the constant.
		stringType := fmt.Sprintf("%s_%s", singularName, enum)
		strEnum = strEnum + fmt.Sprintf("// %s represents the %s `%q`.\n", strcase.ToCamel(stringType), singularName, enum)
		strEnum = strEnum + fmt.Sprintf("\t%s %s = %q\n", strcase.ToCamel(stringType), singularName, enum)

		// Add the enum type to the list of enum types.
		//collectEnumStringTypes[singularTypename] = append(collectEnumStringTypes[singularTypename], enum)
		stringEnums[singularTypename] = append(stringEnums[singularTypename], enum)
	}
	// Close the enum values.
	return strEnum + ")\n", stringEnums
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

// TODO: Remove this function when it's not being used anywhere any more
// writeSchemaTypeDescription writes the description of the given type.
func writeSchemaTypeDescription(name string, s *openapi3.Schema, f *os.File) {
	if s.Description != "" {
		fmt.Fprintf(f, "// %s is %s\n", name, toLowerFirstLetter(strings.ReplaceAll(s.Description, "\n", "\n// ")))
	} else {
		fmt.Fprintf(f, "// %s is the type definition for a %s.\n", name, name)
	}
}

// schemaTypeDescription returns the description of the given type.
func schemaTypeDescription(name string, s *openapi3.Schema) string {
	if s.Description != "" {
		return fmt.Sprintf("// %s is %s\n", name, toLowerFirstLetter(strings.ReplaceAll(s.Description, "\n", "\n// ")))
	}
	return fmt.Sprintf("// %s is the type definition for a %s.\n", name, name)

}
