package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/davecgh/go-spew/spew"
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

		formattedType, typeTpl, enumTpl := writeSchemaType(name, s.Value, "")
		typeCollect = append(typeCollect, typeTpl...)
		enumCollect = append(enumCollect, enumTpl...)

		// TODO: Remove when all types are constructed through structs
		fmt.Fprint(f, formattedType)
	}

	// TODO: Remove, this is only for development
	spew.Dump(typeCollect)
	spew.Dump(enumCollect)

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
func writeSchemaType(name string, s *openapi3.Schema, additionalName string) (string, []TypeTemplate, []EnumTemplate) {
	fmt.Printf("writing type for schema %q -> %s\n", name, s.Type)

	name = strcase.ToCamel(name)
	typeName := strcase.ToCamel(name)
	if additionalName != "" {
		typeName = fmt.Sprintf("%s%s", name, strcase.ToCamel(additionalName))
	}

	types := []TypeTemplate{}
	enumTypes := []EnumTemplate{}
	var typeTpl = TypeTemplate{}
	var typeStr string

	switch ot := getObjectType(s); ot {
	case "string":
		typeTpl.Description = schemaTypeDescription(typeName, s)
		typeTpl.Type = "string"
		typeTpl.Name = name

		// TODO: Remove this line once all types are constructed with the structs
		typeStr = typeStr + schemaTypeDescriptionDeprecated(typeName, s) + fmt.Sprintf("type %s string\n", name)
	case "string_enum":
		strEnum, enums, tt, et := createStringEnum(s, collectEnumStringTypes, name, typeName)
		types = append(types, tt...)
		enumTypes = append(enumTypes, et...)
		// spew.Dump(enums)
		// TODO: Handle string enums with TypeTemplate
		collectEnumStringTypes = enums
		typeStr = fmt.Sprint(strEnum)
	case "integer":
		typeTpl.Description = schemaTypeDescription(typeName, s)
		typeTpl.Type = "int64"
		typeTpl.Name = name

		// TODO: Remove this line once all types are constructed with the structs
		typeStr = typeStr + schemaTypeDescriptionDeprecated(typeName, s) + fmt.Sprintf("type %s int64\n", name)
	case "number":
		typeTpl.Description = schemaTypeDescription(typeName, s)
		typeTpl.Type = "float64"
		typeTpl.Name = name

		// TODO: Remove this line once all types are constructed with the structs
		typeStr = typeStr + schemaTypeDescriptionDeprecated(typeName, s) + fmt.Sprintf("type %s float64\n", name)
	case "boolean":
		typeTpl.Description = schemaTypeDescription(typeName, s)
		typeTpl.Type = "bool"
		typeTpl.Name = name

		// TODO: Remove this line once all types are constructed with the structs
		typeStr = typeStr + schemaTypeDescriptionDeprecated(typeName, s) + fmt.Sprintf("type %s bool\n", name)
	case "array":
		typeTpl.Description = schemaTypeDescription(typeName, s)
		typeTpl.Type = fmt.Sprintf("[]%s", s.Items.Value.Type)
		typeTpl.Name = name
		// spew.Dump(typeTpl)

		// TODO: Remove this line once all types are constructed with the structs
		typeStr = typeStr + schemaTypeDescriptionDeprecated(typeName, s) + fmt.Sprintf("type %s []%s\n", name, s.Items.Value.Type)
	case "object":
		typeObj, tt := createTypeObject(s.Properties, name, typeName, schemaTypeDescription(typeName, s))
		typeTpl = tt

		// TODO: Remove this line once all types are constructed with the structs
		typeStr = typeStr + schemaTypeDescriptionDeprecated(typeName, s) + fmt.Sprint(typeObj)
		// TODO: Handle these differently. The output of these
		// is generally not structs, but constants, variables and string types
		// Iterate over the properties and write the types, if we need to.
		for k, v := range s.Properties {
			if isLocalEnum(v) {
				// TODO: Ignore the TypeTemplate for now
				e, tt, et := writeSchemaType(fmt.Sprintf("%s%s", name, strcase.ToCamel(k)), v.Value, "")
				// TODO: tt doesn't actually contain anything yet as it's mostly enums that I haven't configured yet
				// spew.Dump(tt)
				types = append(types, tt...)
				enumTypes = append(enumTypes, et...)
				typeStr = typeStr + fmt.Sprint(e)
			}

			// TODO: So far this code is never hit with the current openapi spec
			if isLocalObject(v) {
				// TODO: Ignore the TypeTemplate for now
				obj, tt, et := writeSchemaType(fmt.Sprintf("%s%s", name, strcase.ToCamel(k)), v.Value, "")
				types = append(types, tt...)
				enumTypes = append(enumTypes, et...)
				typeStr = typeStr + fmt.Sprint(obj)
			}
		}
	case "one_of":
		typeOneOf, tt, et := createOneOf(s, name, typeName)
		types = append(types, tt...)
		enumTypes = append(enumTypes, et...)
		// spew.Dump(types)
		// TODO: Remove once all types are constructed with structs
		typeStr = fmt.Sprint(typeOneOf)
	case "any_of":
		fmt.Printf("[WARN] TODO: skipping type for %q, since it is a ANYOF\n", name)
	case "all_of":
		fmt.Printf("[WARN] TODO: skipping type for %q, since it is a ALLOF\n", name)
	default:
		fmt.Printf("[WARN] TODO: skipping type for %q, since it is an unknown type\n", name)
	}

	typeStr = typeStr + fmt.Sprintln("")

	types = append(types, typeTpl)

	return typeStr, types, enumTypes
}

// TODO: use the TypeTemplate struct to build these
func createTypeObject(schemas map[string]*openapi3.SchemaRef, name, typeName, description string) (string, TypeTemplate) {
	var typeObj string
	typeObj = fmt.Sprintf("type %s struct {\n", typeName)
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

			typeObj = typeObj + fmt.Sprintf("\t// %s is %s\n", strcase.ToCamel(k), toLowerFirstLetter(strings.ReplaceAll(v.Value.Description, "\n", "\n// ")))
		}

		field.Name = strcase.ToCamel(k)
		field.Type = typeName

		serInfo := fmt.Sprintf("`json:\"%s,omitempty\" yaml:\"%s,omitempty\"`", k, k)
		field.SerializationInfo = serInfo

		// TODO: Remove this once all types are constructed with structs
		typeObj = typeObj + fmt.Sprintf("\t%s %s `json:\"%s,omitempty\" yaml:\"%s,omitempty\"`\n", strcase.ToCamel(k), typeName, k, k)

		fields = append(fields, field)

	}

	typeTpl.Fields = fields

	// TODO: Remove this once all types are constructed with structs
	typeObj = typeObj + "}\n"
	return typeObj, typeTpl
}

func createStringEnum(s *openapi3.Schema, stringEnums map[string][]string, name, typeName string) (string, map[string][]string, []TypeTemplate, []EnumTemplate) {
	var strEnum string
	singularTypename := makeSingular(typeName)
	singularName := makeSingular(name)
	typeTpls := []TypeTemplate{}
	// Make sure we don't redeclare the enum type.
	if _, ok := stringEnums[singularTypename]; !ok {
		strEnum = schemaTypeDescriptionDeprecated(singularTypename, s)

		// Write the enum type.
		strEnum = strEnum + fmt.Sprintf("type %s string\n", singularTypename)

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
	// Close the enum values.
	return strEnum + ")\n", stringEnums, typeTpls, enumTpls
}

func createOneOf(s *openapi3.Schema, name, typeName string) (string, []TypeTemplate, []EnumTemplate) {
	var strOneOf string
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
					Description:       p.Value.Description,
					Name:              strcase.ToCamel(prop),
					Type:              propertyName,
					SerializationInfo: fmt.Sprintf("`json:\"%s,omitempty\" yaml:\"%s,omitempty\"`", prop, prop),
				}

				fields = append(fields, field)

				// TODO: Remove when all types are constructed through structs
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
		t, tt, et := writeSchemaType(name, v.Value, typeName2)
		typeTpls = append(typeTpls, tt...)
		enumTpls = append(enumTpls, et...)
		strOneOf = strOneOf + t
	}

	typeTpl := TypeTemplate{
		Description: schemaTypeDescription(typeName, s),
		Name:        typeName,
		Type:        "struct",
		Fields:      fields,
	}

	typeTpls = append(typeTpls, typeTpl)

	// Now let's create the global oneOf type.
	// Write the type description.
	strOneOf = strOneOf + schemaTypeDescriptionDeprecated(typeName, s)
	strOneOf = strOneOf + fmt.Sprintf("type %s struct {\n", typeName)
	// Iterate over the properties and write the types, if we need to.
	for _, p := range properties {
		strOneOf = strOneOf + p
	}
	// Close the struct.
	return strOneOf + "}\n", typeTpls, enumTpls
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

// TODO: Delete once all types are constructed by the structs
// schemaTypeDescription returns the description of the given type.
func schemaTypeDescriptionDeprecated(name string, s *openapi3.Schema) string {
	if s.Description != "" {
		return fmt.Sprintf("// %s is %s\n", name, toLowerFirstLetter(strings.ReplaceAll(s.Description, "\n", "\n// ")))
	}
	return fmt.Sprintf("// %s is the type definition for a %s.\n", name, name)

}
