package main

import (
	"fmt"
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

//type Types []TypeTemplate

// TypeTemplate holds the information of a type struct
type TypeTemplate struct {
	// Description holds the description of the type
	Description string
	// Name of the type
	Name string
	// Type describes the type of the type (e.g. struct, int64, string)
	Type string
	// Fields holds the information for the field
	Fields []TypeFields
}

// TypeFields holds the information for each type field
type TypeFields struct {
	Description       string
	Name              string
	Type              string
	SerializationInfo string
}

// EnumTemplate holds the information for enum types
type EnumTemplate struct {
	Description string
	Name        string
	ValueType   string
	Value       string
}

// Generate the types file.
func generateTypes(file string, spec *openapi3.T) error {
	f, err := openGeneratedFile(file)
	if err != nil {
		return err
	}
	defer f.Close()

	// Start an empty collection of types
	typeCollect := []TypeTemplate{}

	// Start an empty collection of enum types
	enumCollect := []EnumTemplate{}

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

		typeTpl, enumTpl := writeSchemaType(name, s.Value, "")
		typeCollect = append(typeCollect, typeTpl...)
		enumCollect = append(enumCollect, enumTpl...)
	}

	// TODO: Currently enums are set as variables like this example:
	//
	// var BinRangedoubleTypes = []BinRangedoubleType{
	// 	BinRangedoubleTypeRange,
	// 	BinRangedoubleTypeRangeFrom,
	// 	BinRangedoubleTypeRangeTo,
	// }
	// This approach can be problematic for several reasons
	// The most obvious being that the variable can change its value at any moment
	// The approach to handle enums should be changed.
	//
	// Iterate over all the enum types and add in the slices.
	// We want to ensure we keep the order so the diffs don't look like shit.
	keys = make([]string, 0)
	for k := range collectEnumStringTypes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		// TODO: Remove once all types are constructed through structs
		enums := collectEnumStringTypes[name]

		var enumItems string
		sort.Strings(enums)
		for _, enum := range enums {
			// Most likely, the enum values are strings.
			enumItems = enumItems + fmt.Sprintf("\t%s,\n", strcase.ToCamel(fmt.Sprintf("%s_%s", makeSingular(name), enum)))
		}
		enumVar := fmt.Sprintf("= []%s{\n", makeSingular(name)) + enumItems + "}"

		enumTpl := EnumTemplate{
			Description: fmt.Sprintf("// %s is the collection of all %s values.", makePlural(name), makeSingular(name)),
			Name:        makePlural(name),
			ValueType:   "var",
			Value:       enumVar,
		}

		enumCollect = append(enumCollect, enumTpl)
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
		fmt.Fprintf(f, "%s %s %s\n\n", et.ValueType, et.Name, et.Value)
	}

	return nil
}

// writeSchemaType writes a type definition for the given schema.
// The additional parameter is only used as a suffix for the type name.
// This is mostly for oneOf types.
func writeSchemaType(name string, s *openapi3.Schema, additionalName string) ([]TypeTemplate, []EnumTemplate) {
	fmt.Printf("writing type for schema %q -> %s\n", name, s.Type)

	name = strcase.ToCamel(name)
	typeName := strcase.ToCamel(name)
	if additionalName != "" {
		typeName = fmt.Sprintf("%s%s", name, strcase.ToCamel(additionalName))
	}

	types := []TypeTemplate{}
	enumTypes := []EnumTemplate{}
	var typeTpl = TypeTemplate{}

	switch ot := getObjectType(s); ot {
	case "string":
		typeTpl.Description = schemaTypeDescription(typeName, s)
		typeTpl.Type = "string"
		typeTpl.Name = name
	case "string_enum":
		enums, tt, et := createStringEnum(s, collectEnumStringTypes, name, typeName)
		types = append(types, tt...)
		enumTypes = append(enumTypes, et...)
		collectEnumStringTypes = enums
	case "integer":
		typeTpl.Description = schemaTypeDescription(typeName, s)
		typeTpl.Type = "int64"
		typeTpl.Name = name
	case "number":
		typeTpl.Description = schemaTypeDescription(typeName, s)
		typeTpl.Type = "float64"
		typeTpl.Name = name
	case "boolean":
		typeTpl.Description = schemaTypeDescription(typeName, s)
		typeTpl.Type = "bool"
		typeTpl.Name = name
	case "array":
		typeTpl.Description = schemaTypeDescription(typeName, s)
		typeTpl.Type = fmt.Sprintf("[]%s", s.Items.Value.Type)
		typeTpl.Name = name
	case "object":
		tt := createTypeObject(s.Properties, name, typeName, schemaTypeDescription(typeName, s))
		typeTpl = tt

		// TODO: Handle these differently. The output of these
		// is generally not structs, but constants, variables and string types
		// Iterate over the properties and write the types, if we need to.
		for k, v := range s.Properties {
			if isLocalEnum(v) {
				tt, et := writeSchemaType(fmt.Sprintf("%s%s", name, strcase.ToCamel(k)), v.Value, "")
				types = append(types, tt...)
				enumTypes = append(enumTypes, et...)
			}

			// TODO: So far this code is never hit with the current openapi spec
			if isLocalObject(v) {
				tt, et := writeSchemaType(fmt.Sprintf("%s%s", name, strcase.ToCamel(k)), v.Value, "")
				types = append(types, tt...)
				enumTypes = append(enumTypes, et...)
			}
		}
	case "one_of":
		tt, et := createOneOf(s, name, typeName)
		types = append(types, tt...)
		enumTypes = append(enumTypes, et...)
	case "any_of":
		fmt.Printf("[WARN] TODO: skipping type for %q, since it is a ANYOF\n", name)
	case "all_of":
		fmt.Printf("[WARN] TODO: skipping type for %q, since it is a ALLOF\n", name)
	default:
		fmt.Printf("[WARN] TODO: skipping type for %q, since it is an unknown type\n", name)
	}

	types = append(types, typeTpl)

	return types, enumTypes
}

// TODO: use the TypeTemplate struct to build these
func createTypeObject(schemas map[string]*openapi3.SchemaRef, name, typeName, description string) TypeTemplate {
	typeTpl := TypeTemplate{
		Description: description,
		Name:        typeName,
		Type:        "struct",
	}

	// We want to ensure we keep the order so the diffs don't look like shit.
	keys := make([]string, 0)
	for k := range schemas {
		keys = append(keys, k)
	}
	fields := []TypeFields{}
	sort.Strings(keys)
	for _, k := range keys {
		v := schemas[k]
		// Check if we need to generate a type for this type.
		typeName := convertToValidGoType(k, v)

		if isLocalEnum(v) {
			typeName = fmt.Sprintf("%s%s", name, strcase.ToCamel(k))
		}

		// TODO: So far this code is never hit with the current openapi spec
		if isLocalObject(v) {
			typeName = fmt.Sprintf("%s%s", name, strcase.ToCamel(k))
		}

		field := TypeFields{}
		if v.Value.Description != "" {
			desc := fmt.Sprintf("// %s is %s", strcase.ToCamel(k), toLowerFirstLetter(strings.ReplaceAll(v.Value.Description, "\n", "\n// ")))
			field.Description = desc
		}

		field.Name = strcase.ToCamel(k)
		field.Type = typeName
		serInfo := fmt.Sprintf("`json:\"%s,omitempty\" yaml:\"%s,omitempty\"`", k, k)
		field.SerializationInfo = serInfo
		fields = append(fields, field)

	}
	typeTpl.Fields = fields

	return typeTpl
}

func createStringEnum(s *openapi3.Schema, stringEnums map[string][]string, name, typeName string) (map[string][]string, []TypeTemplate, []EnumTemplate) {
	singularTypename := makeSingular(typeName)
	singularName := makeSingular(name)
	typeTpls := []TypeTemplate{}
	// Make sure we don't redeclare the enum type.
	if _, ok := stringEnums[singularTypename]; !ok {
		typeTpl := TypeTemplate{
			Description: schemaTypeDescription(singularName, s),
			Name:        singularTypename,
			Type:        "string",
		}

		typeTpls = append(typeTpls, typeTpl)

		stringEnums[singularTypename] = []string{}
	}

	// Define the enum values.
	enumTpls := []EnumTemplate{}
	for _, v := range s.Enum {
		// Most likely, the enum values are strings.
		enum, ok := v.(string)
		if !ok {
			fmt.Printf("[WARN] TODO: enum value is not a string for %q -> %#v\n", name, v)
			continue
		}
		// Write the description of the constant.
		stringType := fmt.Sprintf("%s_%s", singularName, enum)

		enumTpl := EnumTemplate{
			Description: fmt.Sprintf("// %s represents the %s `%q`.", strcase.ToCamel(stringType), singularName, enum),
			Name:        strcase.ToCamel(stringType),
			ValueType:   "const",
			Value:       fmt.Sprintf("%s = %q", singularName, enum),
		}

		enumTpls = append(enumTpls, enumTpl)

		// Add the enum type to the list of enum types.
		stringEnums[singularTypename] = append(stringEnums[singularTypename], enum)
	}

	return stringEnums, typeTpls, enumTpls
}

func createOneOf(s *openapi3.Schema, name, typeName string) ([]TypeTemplate, []EnumTemplate) {
	var properties []string
	enumTpls := []EnumTemplate{}
	typeTpls := []TypeTemplate{}
	fields := []TypeFields{}
	for _, v := range s.OneOf {
		// We want to iterate over the properties of the embedded object
		// and find the type that is a string.
		var typeName2 string

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
			// Avoids duplication for every enum
			if !containsMatchFirstWord(properties, propertyString) {
				// Construct TypeFields
				field := TypeFields{
					Description:       schemaTypeDescription(strcase.ToCamel(prop), p.Value),
					Name:              strcase.ToCamel(prop),
					Type:              propertyName,
					SerializationInfo: fmt.Sprintf("`json:\"%s,omitempty\" yaml:\"%s,omitempty\"`", prop, prop),
				}
				fields = append(fields, field)

				// TODO: Is this needed?
				properties = append(properties, propertyString)
			}

			if p.Value.Enum != nil {
				// We want to get the enum value.
				// Make sure there is only one.
				if len(p.Value.Enum) != 1 {
					fmt.Printf("[WARN] TODO: oneOf for %q -> %q enum %#v\n", name, prop, p.Value.Enum)
					continue
				}

				typeName2 = strcase.ToCamel(p.Value.Enum[0].(string))
			}
		}

		// TODO: This is the only place that has an "additional name" at the end
		tt, et := writeSchemaType(name, v.Value, typeName2)
		typeTpls = append(typeTpls, tt...)
		enumTpls = append(enumTpls, et...)
	}

	typeTpl := TypeTemplate{
		Description: schemaTypeDescription(typeName, s),
		Name:        typeName,
		Type:        "struct",
		Fields:      fields,
	}
	typeTpls = append(typeTpls, typeTpl)

	return typeTpls, enumTpls
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

// schemaTypeDescription returns the description of the given type.
func schemaTypeDescription(name string, s *openapi3.Schema) string {
	if s.Description != "" {
		return fmt.Sprintf("// %s is %s", name, toLowerFirstLetter(strings.ReplaceAll(s.Description, "\n", "\n// ")))
	}
	return fmt.Sprintf("// %s is the type definition for a %s.", name, name)

}
