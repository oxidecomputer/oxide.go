// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
// For now they are being collected to make sure they
// are not duplicated in createStringEnum()
var collectEnumStringTypes = enumStringTypes()

func enumStringTypes() map[string][]string {
	return map[string][]string{}
}

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

// ValidationTemplate holds information about the fields that
// need to be validated
type ValidationTemplate struct {
	RequiredObjects []string
	RequiredStrings []string
	RequiredNums    []string
	AssociatedType  string
}

// Generate the types file.
func generateTypes(file string, spec *openapi3.T) error {
	f, err := openGeneratedFile(file)
	if err != nil {
		return err
	}
	defer f.Close()

	typeCollection, enumCollection := constructTypes(spec.Components.Schemas)
	enumCollection = append(enumCollection, constructEnums(collectEnumStringTypes)...)
	typeCollection = append(typeCollection, constructParamTypes(spec.Paths.Map())...)
	v := constructParamValidation(spec.Paths.Map())

	writeTypes(f, typeCollection, v, enumCollection)

	return nil
}

func constructParamTypes(paths map[string]*openapi3.PathItem) []TypeTemplate {
	paramTypes := make([]TypeTemplate, 0)

	keys := make([]string, 0)
	for k := range paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, path := range keys {
		p := paths[path]
		if p.Ref != "" {
			fmt.Printf("[WARN] TODO: skipping path for %q, since it is a reference\n", path)
			continue
		}
		ops := p.Operations()
		keys := make([]string, 0)
		for k := range ops {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, op := range keys {
			o := ops[op]
			requiredFields := ""

			// Some required fields are defined in vendor extensions
			for k, v := range o.Extensions {
				if k == "x-dropshot-pagination" {
					for i, j := range v.(map[string]interface{}) {
						if i == "required" {
							values, ok := j.([]interface{})
							if ok {
								for _, field := range values {
									str, ok := field.(string)
									if ok {
										requiredFields = requiredFields + fmt.Sprintf("\n// - %v", strcase.ToCamel(str))
									}
								}
							}
						}
					}
				}
			}

			if len(o.Parameters) > 0 || o.RequestBody != nil {
				paramsTypeName := strcase.ToCamel(o.OperationID) + "Params"
				paramsTpl := TypeTemplate{
					Type: "struct",
					Name: paramsTypeName,
				}

				fields := make([]TypeFields, 0)
				for _, p := range o.Parameters {
					if p.Ref != "" {
						fmt.Printf("[WARN] TODO: skipping parameter for %q, since it is a reference\n", p.Value.Name)
						continue
					}

					paramName := strcase.ToCamel(p.Value.Name)
					field := TypeFields{
						Name: paramName,
						Type: convertToValidGoType("", "", p.Value.Schema),
					}

					if p.Value.Required {
						requiredFields = requiredFields + fmt.Sprintf("\n// - %s", paramName)
					}

					serInfo := fmt.Sprintf("`json:\"%s,omitempty\" yaml:\"%s,omitempty\"`", p.Value.Name, p.Value.Name)
					field.SerializationInfo = serInfo

					fields = append(fields, field)
				}
				if o.RequestBody != nil {
					var field TypeFields
					// The Nexus API spec only has a single value for content, so we can safely
					// break when a condition is met
					for mt, r := range o.RequestBody.Value.Content {
						// TODO: Handle other mime types in a more idiomatic way
						if mt != "application/json" {
							field = TypeFields{
								Name:              "Body",
								Type:              "io.Reader",
								SerializationInfo: "`json:\"body,omitempty\" yaml:\"body,omitempty\"`",
							}
							break
						}

						field = TypeFields{
							Name:              "Body",
							Type:              "*" + convertToValidGoType("", "", r.Schema),
							SerializationInfo: "`json:\"body,omitempty\" yaml:\"body,omitempty\"`",
						}
					}
					// Body is always a required field
					requiredFields = requiredFields + "\n// - Body"
					fields = append(fields, field)
				}
				paramsTpl.Fields = fields

				description := "// " + paramsTypeName + " is the request parameters for " +
					strcase.ToCamel(o.OperationID)
				if requiredFields != "" {
					description = description + "\n//\n// Required fields:" + requiredFields
				}
				paramsTpl.Description = description
				paramTypes = append(paramTypes, paramsTpl)
			}
		}

	}

	return paramTypes
}

func constructParamValidation(paths map[string]*openapi3.PathItem) []ValidationTemplate {
	validationMethods := make([]ValidationTemplate, 0)

	keys := make([]string, 0)
	for k := range paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, path := range keys {
		p := paths[path]
		if p.Ref != "" {
			fmt.Printf("[WARN] TODO: skipping path for %q, since it is a reference\n", path)
			continue
		}
		ops := p.Operations()
		keys := make([]string, 0)
		for k := range ops {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, op := range keys {
			o := ops[op]
			if len(o.Parameters) > 0 || o.RequestBody != nil {
				paramsTypeName := strcase.ToCamel(o.OperationID) + "Params"

				validationTpl := ValidationTemplate{
					AssociatedType: paramsTypeName,
				}

				for _, p := range o.Parameters {
					if !p.Value.Required {
						continue
					}

					paramName := strcase.ToCamel(p.Value.Name)

					if p.Value.Schema.Value.Type.Is("integer") {
						validationTpl.RequiredNums = append(validationTpl.RequiredNums, paramName)
						continue
					}

					validationTpl.RequiredStrings = append(validationTpl.RequiredStrings, paramName)
				}

				if o.RequestBody != nil {
					// If an endpoint has a body, our API requires it, so we can safely add it.
					validationTpl.RequiredObjects = append(validationTpl.RequiredObjects, "Body")

				}
				validationMethods = append(validationMethods, validationTpl)
			}
		}

	}

	return validationMethods
}

// constructTypes takes the types collected from several parts of the spec and constructs
// the templates
func constructTypes(schemas openapi3.Schemas) ([]TypeTemplate, []EnumTemplate) {
	typeCollection := make([]TypeTemplate, 0)
	enumCollection := make([]EnumTemplate, 0)

	keys := make([]string, 0)
	for k := range schemas {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		s := schemas[name]
		if s.Ref != "" {
			fmt.Printf("[WARN] TODO: skipping type for %q, since it is a reference\n", name)
			continue
		}

		if name == "DatumType" {
			fmt.Printf("[WARN] TODO: skipping type for %q, since it is a duplicate\n", name)
			continue
		}

		// Set name as a valid Go type name
		name = strcase.ToCamel(name)
		typeTpl, enumTpl := populateTypeTemplates(name, s.Value, "")
		typeCollection = append(typeCollection, typeTpl...)
		enumCollection = append(enumCollection, enumTpl...)
	}

	return typeCollection, enumCollection
}

// constructEnums takes the enums collected from several parts of the spec and constructs
// the templates
func constructEnums(enumStrCollection map[string][]string) []EnumTemplate {
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

	enumCollection := make([]EnumTemplate, 0)

	// Iterate over all the enum types and add in the slices.
	// We want to ensure we keep the order
	keys := make([]string, 0)
	for k := range enumStrCollection {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		// TODO: Remove once all types are constructed through structs
		enums := enumStrCollection[name]

		var enumItems string
		sort.Strings(enums)
		for _, enum := range enums {
			// Most likely, the enum values are strings.
			enumItems = enumItems + fmt.Sprintf("\t%s,\n", strcase.ToCamel(fmt.Sprintf("%s_%s", name, enum)))
		}

		if enumItems == "" {
			continue
		}

		enumVar := fmt.Sprintf("= []%s{\n", name) + enumItems + "}"

		varName := name + "Collection"
		enumTpl := EnumTemplate{
			Description: fmt.Sprintf("// %s is the collection of all %s values.", varName, name),
			Name:        varName,
			ValueType:   "var",
			Value:       enumVar,
		}

		enumCollection = append(enumCollection, enumTpl)
	}

	return enumCollection
}

// writeTypes iterates over the templates, constructs the different types and writes to file
func writeTypes(f *os.File, typeCollection []TypeTemplate, typeValidationCollection []ValidationTemplate, enumCollection []EnumTemplate) {
	// Write all collected types to file
	for _, tt := range typeCollection {
		// If an empty template manages to get through, ignore it.
		// if there is a weirdly constructed template, then let it get through
		// so it's evident to us.
		if tt.Name == "" && tt.Type == "" && tt.Description == "" {
			continue
		}

		fmt.Fprintf(f, "%s\n", splitDocString(tt.Description))
		fmt.Fprintf(f, "type %s %s", tt.Name, tt.Type)
		if tt.Fields != nil {
			fmt.Fprint(f, " {\n")
			for _, ft := range tt.Fields {
				if ft.Description != "" {
					fmt.Fprintf(f, "\t%s\n", splitDocString(ft.Description))
				}
				fmt.Fprintf(f, "\t%s %s %s\n", ft.Name, ft.Type, ft.SerializationInfo)
			}
			fmt.Fprint(f, "}\n")
		}
		fmt.Fprint(f, "\n")
	}

	// Write all collected validation methods to file
	for _, vm := range typeValidationCollection {
		if vm.AssociatedType == "" {
			continue
		}

		fmt.Fprintf(f, "// Validate verifies all required fields for %s are set\n", vm.AssociatedType)
		fmt.Fprintf(f, "func (p *%s) Validate() error {\n", vm.AssociatedType)
		fmt.Fprintln(f, "v := new(Validator)")
		for _, o := range vm.RequiredObjects {
			fmt.Fprintf(f, "v.HasRequiredObj(p.%s, \"%s\")\n", o, o)
		}
		for _, s := range vm.RequiredStrings {
			fmt.Fprintf(f, "v.HasRequiredStr(string(p.%s), \"%s\")\n", s, s)
		}
		for _, i := range vm.RequiredNums {
			fmt.Fprintf(f, "v.HasRequiredNum(p.%s, \"%s\")\n", i, i)
		}
		fmt.Fprintln(f, "if !v.IsValid() {")
		// Unfortunately I have to craft the following line this way as I get
		// unwanted newlines otherwise :(
		n := `\n`
		fmt.Fprintf(f, "return fmt.Errorf(\"validation error:%v%v\", v.Error())", n, `%v`)
		fmt.Fprintln(f, "}")
		fmt.Fprintln(f, "return nil")
		fmt.Fprintln(f, "}")
	}

	// Write all collected enums to file
	for _, et := range enumCollection {
		if et.Name == "" {
			continue
		}

		fmt.Fprintf(f, "%s\n", splitDocString(et.Description))
		fmt.Fprintf(f, "%s %s %s\n\n", et.ValueType, et.Name, et.Value)
	}
}

// populateTypeTemplates populates the template of a type definition for the given schema.
// The additional parameter is only used as a suffix for the type name.
// This is mostly for oneOf types.
func populateTypeTemplates(name string, s *openapi3.Schema, enumFieldName string) ([]TypeTemplate, []EnumTemplate) {
	typeName := name

	// Type name will change for each enum type
	if enumFieldName != "" {
		typeName = fmt.Sprintf("%s%s", name, strcase.ToCamel(enumFieldName))
	}

	types := make([]TypeTemplate, 0)
	enumTypes := make([]EnumTemplate, 0)
	typeTpl := TypeTemplate{}

	// TODO: remove workaround once no more type objects are empty
	if sliceContains(emptyTypes(), name) {
		bgpOT := getObjectType(s)
		if bgpOT != "" {
			panic("[ERROR] " + name + " is no longer an empty type. Remove workaround in exceptions.go")
		}
		s.Type = &openapi3.Types{"string"}
	}

	switch ot := getObjectType(s); ot {
	case "string_enum":
		enums, tt, et := createStringEnum(s, collectEnumStringTypes, name, typeName)
		types = append(types, tt...)
		enumTypes = append(enumTypes, et...)
		collectEnumStringTypes = enums
	case "string", "*bool", "int", "int8", "int16", "int32", "int64", "uint", "uint8",
		"uint16", "uint32", "uint64", "uintptr", "float32", "float64":
		typeTpl.Description = formatTypeDescription(typeName, s)
		typeTpl.Type = ot
		typeTpl.Name = typeName
	case "array":
		typeTpl.Description = formatTypeDescription(typeName, s)
		typeTpl.Type = fmt.Sprintf("[]%s", s.Items.Value.Type)
		typeTpl.Name = typeName
	case "object":
		typeTpl = createTypeObject(s, name, typeName, formatTypeDescription(typeName, s))

		// Iterate over the properties and append the types, if we need to.
		for k, v := range s.Properties {
			if isLocalEnum(v) {
				tt, et := populateTypeTemplates(fmt.Sprintf("%s%s", name, strcase.ToCamel(k)), v.Value, "")
				types = append(types, tt...)
				enumTypes = append(enumTypes, et...)
			}

			// TODO: So far this code is never hit with the current openapi spec
			if isLocalObject(v) {
				tt, et := populateTypeTemplates(fmt.Sprintf("%s%s", name, strcase.ToCamel(k)), v.Value, "")
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
		tt := createAllOf(s, collectEnumStringTypes, name, typeName)
		types = append(types, tt...)

	default:
		fmt.Printf("[WARN] TODO: skipping type for %q, since it is an unknown type\n", name)
	}

	// enums are handled separately, so an empty template would be returned
	if typeTpl.Name != "" {
		types = append(types, typeTpl)
	}

	return types, enumTypes
}

func createTypeObject(schema *openapi3.Schema, name, typeName, description string) TypeTemplate {
	// TODO: Create types out of the schemas instead of plucking them out of the objects
	// will leave this for another PR, because the yak shaving is getting ridiculous.
	// Tracked -> https://github.com/oxidecomputer/oxide.go/issues/110
	//
	// This particular type was being defined here and also in createOneOf()
	if typeName == "ExpectedDigest" {
		return TypeTemplate{}
	}

	typeTpl := TypeTemplate{
		Name: typeName,
		Type: "struct",
	}

	schemas := schema.Properties
	// We want to ensure we keep the order
	keys := make([]string, 0)
	for k := range schemas {
		keys = append(keys, k)
	}
	fields := []TypeFields{}
	sort.Strings(keys)
	for _, k := range keys {
		v := schemas[k]
		// Check if we need to generate a type for this type.
		typeName := convertToValidGoType(k, typeName, v)

		if isLocalEnum(v) {
			typeName = fmt.Sprintf("%s%s", name, strcase.ToCamel(k))
		}

		// TODO: So far this code is never hit with the current openapi spec
		if isLocalObject(v) {
			typeName = fmt.Sprintf("%s%s", name, strcase.ToCamel(k))
		}

		// When additional properties is set and is type array, it means
		// the type will be a map.
		// TODO correctness: Currently our API spec does not specify what
		// type the key will be, so we set it to string to avoid errors.
		// If the type of the key is defined in our spec in the future,
		// this should be changed to reflect that type.
		if isObjectArray(v) {
			typeName = fmt.Sprintf("map[string][]%s", typeName)
		}

		// If a type is nullable we'll want a pointer
		if sliceContains(nullable(), typeName) {
			typeName = fmt.Sprintf("*%s", typeName)
		}

		field := TypeFields{}
		if v.Value.Description != "" {
			desc := fmt.Sprintf("// %s is %s", strcase.ToCamel(k), toLowerFirstLetter(strings.ReplaceAll(v.Value.Description, "\n", "\n// ")))
			field.Description = desc
		}

		field.Name = strcase.ToCamel(k)
		field.Type = typeName

		// TODO: Set omitzero on all types.
		// https://github.com/oxidecomputer/oxide.go/issues/290
		serInfo := fmt.Sprintf("`json:\"%s,omitempty\" yaml:\"%s,omitempty\"`", k, k)
		if isNullableArray(v) {
			serInfo = fmt.Sprintf("`json:\"%s\" yaml:\"%s\"`", k, k)
		} else if sliceContains(omitzeroTypes(), typeName) {
			serInfo = fmt.Sprintf("`json:\"%s,omitzero\" yaml:\"%s,omitzero\"`", k, k)
		}

		field.SerializationInfo = serInfo

		fields = append(fields, field)

	}
	typeTpl.Fields = fields

	if len(schema.Required) > 0 {
		description = description + "\n//\n// Required fields:"
		for _, r := range schema.Required {
			description = description + fmt.Sprintf("\n// - %s", strcase.ToCamel(r))
		}
	}
	typeTpl.Description = description

	return typeTpl
}

func createStringEnum(s *openapi3.Schema, stringEnums map[string][]string, name, typeName string) (map[string][]string, []TypeTemplate, []EnumTemplate) {
	typeTpls := make([]TypeTemplate, 0)

	// Make sure we don't redeclare the enum type.
	if _, ok := stringEnums[typeName]; !ok {
		typeTpl := TypeTemplate{
			Description: formatTypeDescription(name, s),
			Name:        typeName,
			Type:        "string",
		}

		typeTpls = append(typeTpls, typeTpl)
		stringEnums[typeName] = []string{}
	}

	enumTpls := make([]EnumTemplate, 0)
	for _, v := range s.Enum {
		// TODO: Handle different enum types more gracefully
		// Most likely, the enum values are strings.
		enum, ok := v.(string)
		if !ok {
			fmt.Printf("[WARN] TODO: enum value is not a string for %q -> %#v\n", name, v)
			continue
		}
		snakeCaseTypeName := fmt.Sprintf("%s_%s", name, enum)

		enumTpl := EnumTemplate{
			Description: fmt.Sprintf("// %s represents the %s `%q`.", strcase.ToCamel(snakeCaseTypeName), name, enum),
			Name:        strcase.ToCamel(snakeCaseTypeName),
			ValueType:   "const",
			Value:       fmt.Sprintf("%s = %q", name, enum),
		}

		enumTpls = append(enumTpls, enumTpl)

		// Add the enum type to the list of enum types.
		stringEnums[typeName] = append(stringEnums[typeName], enum)
	}

	return stringEnums, typeTpls, enumTpls
}

// TODO: For now AllOf values are treated as interfaces. This way you can pass whichever
// of the struct types you need like this:
//
//	ipRange := oxide.Ipv4Range{
//		 First: "172.20.15.240",
//		 Last:  "172.20.15.250",
//	}
//
// body := oxide.IpRange(ipRange)
// resp, err := client.IpPoolRangeAdd("mypool", &body)
//
// Probably not the best approach, but will leave them this way until I come up with
// a more idiomatic solution. Keep an eye out on this one to refine.
func createAllOf(s *openapi3.Schema, stringEnums map[string][]string, name, typeName string) []TypeTemplate {
	typeTpls := make([]TypeTemplate, 0)

	// Make sure we don't redeclare the enum type.
	if _, ok := stringEnums[typeName]; !ok {
		typeTpl := TypeTemplate{
			Description: formatTypeDescription(name, s),
			Name:        typeName,
			Type:        "interface{}",
		}

		// TODO: See above about making a more idiomatic approach, this is a small workaround
		// until https://github.com/oxidecomputer/oxide.go/issues/67 is done
		if typeName == "NameOrId" {
			typeTpl.Type = "string"
		}

		typeTpls = append(typeTpls, typeTpl)

		stringEnums[typeName] = []string{}
	}

	return typeTpls
}

func createOneOf(s *openapi3.Schema, name, typeName string) ([]TypeTemplate, []EnumTemplate) {
	var parsedProperties []string
	var properties []string
	var genericTypes []string
	enumTpls := make([]EnumTemplate, 0)
	typeTpls := make([]TypeTemplate, 0)
	fields := make([]TypeFields, 0)
	for _, v := range s.OneOf {
		// Iterate over all the schema components in the spec and write the types.
		// We want to ensure we keep the order.
		keys := make([]string, 0)
		for k := range v.Value.Properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, prop := range keys {
			p := v.Value.Properties[prop]
			// We want to collect all the unique properties to create our global oneOf type.
			propertyType := convertToValidGoType(prop, typeName, p)
			properties = append(properties, prop+"="+propertyType)
		}
	}

	// When dealing with oneOf sometimes property types will not be the same, we want to
	// catch these to set them as "any" when we generate the type.
	typeKeys := []string{}
	// First we gather all unique properties
	for _, v := range properties {
		parts := strings.Split(v, "=")
		key := parts[0]
		if !sliceContains(typeKeys, key) {
			typeKeys = append(typeKeys, key)
		}
	}

	// For each of the properties above we gather all possible types
	// and gather all of those that are not. We will be setting those
	// as a generic type
	for _, k := range typeKeys {
		values := []string{}
		for _, v := range properties {
			parts := strings.Split(v, "=")
			key := parts[0]
			value := parts[1]
			if key == k {
				values = append(values, value)
			}
		}

		if !allItemsAreSame(values) {
			genericTypes = append(genericTypes, k)
		}
	}

	for _, v := range s.OneOf {
		// We want to iterate over the properties of the embedded object
		// and find the type that is a string.
		var enumFieldName string

		// Iterate over all the schema components in the spec and write the types.
		// We want to ensure we keep the order.
		keys := make([]string, 0)
		for k := range v.Value.Properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, prop := range keys {
			p := v.Value.Properties[prop]
			// We want to collect all the unique properties to create our global oneOf type.
			propertyType := convertToValidGoType(prop, typeName, p)

			// Check if we have an enum in order to use the corresponding type instead of
			// "string"
			if propertyType == "string" && len(p.Value.Enum) != 0 {
				propertyType = typeName + strcase.ToCamel(prop)
			}

			propertyName := strcase.ToCamel(prop)

			// Avoids duplication for every enum
			if !containsMatchFirstWord(parsedProperties, propertyName) {
				field := TypeFields{
					Description:       formatTypeDescription(propertyName, p.Value),
					Name:              propertyName,
					Type:              propertyType,
					SerializationInfo: fmt.Sprintf("`json:\"%s,omitempty\" yaml:\"%s,omitempty\"`", prop, prop),
				}

				// We set the type of a field as "any" if every element of the oneOf property isn't the same
				if sliceContains(genericTypes, prop) {
					field.Type = "any"
				}

				fields = append(fields, field)

				parsedProperties = append(parsedProperties, propertyName)
			}

			if p.Value.Enum != nil {
				// We want to get the enum value.
				// Make sure there is only one.
				if len(p.Value.Enum) != 1 {
					fmt.Printf("[WARN] TODO: oneOf for %q -> %q enum %#v\n", name, prop, p.Value.Enum)
					continue
				}

				enumFieldName = strcase.ToCamel(p.Value.Enum[0].(string))
			}

			// Enums can appear in a valid OpenAPI spec as a OneOf without necessarily
			// being identified as such. If we find an object with a single property
			// nested inside a OneOf we will assume this is an enum and modify the name of
			// the struct that will be created out of this object.
			// e.g. https://github.com/oxidecomputer/omicron/blob/158c0b205f23772dc6c4c97633fd1769cc0e00d4/openapi/nexus.json#L18637-L18682
			if len(keys) == 1 && p.Value.Enum == nil {
				enumFieldName = propertyName
			}
		}

		// TODO: This is the only place that has an "additional name" at the end
		// TODO: This is where the "allOf" is being detected
		tt, et := populateTypeTemplates(name, v.Value, enumFieldName)
		typeTpls = append(typeTpls, tt...)
		enumTpls = append(enumTpls, et...)
	}

	// TODO: For now AllOf values within a OneOf are treated as enums
	// because that's how they are being used. Keep an eye out if this
	// changes
	for _, v := range s.OneOf {
		if v.Value.AllOf != nil {
			return typeTpls, enumTpls
		}
	}

	// Make sure to only create structs if the oneOf is not a replacement for enums on the API spec
	if len(fields) > 0 {
		typeTpl := TypeTemplate{
			Description: formatTypeDescription(typeName, s),
			Name:        typeName,
			Type:        "struct",
			Fields:      fields,
		}
		typeTpls = append(typeTpls, typeTpl)
	}
	return typeTpls, enumTpls
}

func getObjectType(s *openapi3.Schema) string {
	// TODO: Support enums of other types
	if s.Type.Is("string") && len(s.Enum) > 0 {
		return "string_enum"
	}

	if s.Type.Is("integer") {
		if isNumericType(s.Format) {
			return s.Format
		}
		return "int"
	}

	if s.Type.Is("number") {
		if isNumericType(s.Format) {
			return s.Format
		}
		return "float64"
	}

	if s.Type.Is("boolean") {
		return "*bool"
	}

	if s.Type != nil {
		t := s.Type.Slice()
		// Our API only supports a single type per object
		return string(t[0])
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

// formatTypeDescription returns the description of the given type.
func formatTypeDescription(name string, s *openapi3.Schema) string {
	if s.Description != "" {
		return fmt.Sprintf("// %s is %s", name, toLowerFirstLetter(strings.ReplaceAll(s.Description, "\n", "\n// ")))
	}
	return fmt.Sprintf("// %s is the type definition for a %s.", name, name)
}
