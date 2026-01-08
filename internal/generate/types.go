// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"slices"
	"sort"
	"strings"
	"text/template"

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

var (
	typeTemplate = template.Must(
		template.New("type.go.tpl").
			Funcs(template.FuncMap{
				"splitDocString": splitDocString,
			}).
			ParseFiles("./templates/type.go.tpl"),
	)
	enumTemplate = template.Must(
		template.New("enum.go.tpl").
			Funcs(template.FuncMap{"splitDocString": splitDocString}).
			ParseFiles("./templates/enum.go.tpl"),
	)
	validationTemplate = template.Must(
		template.ParseFiles("./templates/validation.go.tpl"),
	)
)

// renderTemplate executes a template with the given data and returns the result.
func renderTemplate(tmpl *template.Template, data any) string {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		panic(err)
	}
	return buf.String()
}

// TypeTemplate holds the information of a type struct
type TypeTemplate struct {
	// Description holds the description of the type
	Description string
	// Name of the type
	Name string
	// Type describes the type of the type (e.g. struct, int64, string, interface)
	Type string
	// Fields holds the information for the field
	Fields []TypeField

	// OneOfMarker is the marker method for oneOf interface types and their implementations (e.g., "isFieldValueValue")
	OneOfMarker string

	// OneOfInfo holds information for generating oneOf code including custom UnmarshalJSON methods
	OneOfInfo *OneOfInfo
}

// Discriminator holds information about a oneOf discriminator field.
type Discriminator struct {
	// Field is the Go struct field name (e.g., "Type")
	Field string
	// Key is the JSON key (e.g., "type")
	Key string
	// Type is the Go type name (e.g., "FieldValueType")
	Type string
}

// newDiscriminator creates a Discriminator from the JSON key and parent type name.
func newDiscriminator(key, typeName string) *Discriminator {
	return &Discriminator{
		Field: strcase.ToCamel(key),
		Key:   key,
		Type:  typeName + strcase.ToCamel(key),
	}
}

// OneOfInfo holds information needed to generate code for oneOf types,
// including custom UnmarshalJSON methods for discriminator-based unmarshaling.
type OneOfInfo struct {
	// Discriminator holds information about the discriminator field
	Discriminator Discriminator
	// ValueField is the struct field name that holds the interface value (e.g., "Value")
	ValueField string
	// ValueKey is the JSON key for the value field (e.g., "value")
	ValueKey string
	// ValueType is the Go type name for the value field (e.g., "diskSourceVariant")
	ValueType string
	// Variants maps discriminator enum values to their implementation types
	Variants []OneOfVariant
}

// OneOfVariant represents a single variant in a oneOf discriminated union.
type OneOfVariant struct {
	// EnumValue is the discriminator enum constant (e.g., "FieldValueTypeString")
	EnumValue string
	// ImplType is the concrete type to unmarshal into (e.g., "FieldValueString")
	ImplType string
}

// Render renders the TypeTemplate to a Go type.
func (t TypeTemplate) Render() string {
	return renderTemplate(typeTemplate, t)
}

// TypeField holds the information for each type field.
type TypeField struct {
	Schema     *openapi3.SchemaRef
	Name       string
	Type       string
	MarshalKey string
	Required   bool

	// FallbackDescription generates a generic description for the field when the Schema doesn't have one.
	// TODO: Drop this, since generated descriptions don't contain useful information.
	FallbackDescription bool

	// OmitDirective overrides the derived omit directive for the field.
	// TODO: Drop this; we should set omit directives consistently rather than using overrides.
	OmitDirective string
}

// Description returns the formatted description comment for this field.
func (f TypeField) Description() string {
	if f.Schema == nil {
		return ""
	}
	if f.Schema.Value.Description != "" {
		desc := fmt.Sprintf("// %s is %s", f.Name, toLowerFirstLetter(
			strings.ReplaceAll(f.Schema.Value.Description, "\n", "\n// ")))
		return splitDocString(desc)
	}
	if f.FallbackDescription {
		return splitDocString(fmt.Sprintf("// %s is the type definition for a %s.", f.Name, f.Name))
	}
	return ""
}

// StructTag returns the JSON/YAML struct tags for this field.
// Configure json/yaml struct tags. By default, omit empty/zero
// values, but retain them for required fields.
//
// TODO: Use `omitzero` rather than `omitempty` on all relevant
// fields: https://github.com/oxidecomputer/oxide.go/issues/290
func (f TypeField) StructTag() string {
	var omitDirective string
	switch {
	case f.OmitDirective != "":
		omitDirective = f.OmitDirective
	case f.Required:
		omitDirective = ""
	case f.Schema == nil:
		omitDirective = "omitempty"
	case isNullableArray(f.Schema):
		omitDirective = ""
	case slices.Contains(omitzeroTypes(), f.Type):
		omitDirective = "omitzero"
	case isObjectType(f.Schema):
		// Use omitzero for object/struct types since omitempty doesn't work for structs in Go.
		// A zero-value struct is not considered "empty" by omitempty.
		omitDirective = "omitzero"
	default:
		omitDirective = "omitempty"
	}

	tagValue := f.MarshalKey
	if omitDirective != "" {
		tagValue = f.MarshalKey + "," + omitDirective
	}

	return fmt.Sprintf("`json:\"%s\" yaml:\"%s\"`", tagValue, tagValue)
}

// IsPointer returns whether this field should be a pointer type.
//
// Note: Primitive type pointer logic (int, bool, time) is handled in
// schemaValueToGoType() because those types can appear in nested contexts
// (map values, array items) that don't go through TypeField.
func (f TypeField) IsPointer() bool {
	if f.Schema == nil {
		return false
	}

	v := f.Schema.Value

	// Required + nullable fields should be pointers (Omicron API pattern):
	// they can be set to a null value, but they must not be omitted.
	// The SDK presents these fields as optional and serializes them to
	// `null` if not provided.
	if f.Required && v.Nullable {
		return true
	}

	// Check hardcoded nullable exceptions (upstream API workarounds)
	if slices.Contains(nullable(), f.Type) {
		return true
	}

	return false
}

// GoType returns the Go type for this field, with pointer prefix if needed.
func (f TypeField) GoType() string {
	if f.IsPointer() && !strings.HasPrefix(f.Type, "*") {
		return "*" + f.Type
	}
	return f.Type
}

// EnumTemplate holds the information for enum types.
type EnumTemplate struct {
	Description string
	Name        string
	ValueType   string
	Value       string
}

// Render renders the EnumTemplate as var/const enum item.
func (e EnumTemplate) Render() string {
	return renderTemplate(enumTemplate, e)
}

// ValidationTemplate holds information about the fields that
// need to be validated.
type ValidationTemplate struct {
	RequiredObjects []string
	RequiredStrings []string
	RequiredNums    []string
	AssociatedType  string
}

// Render renders the ValidationTemplate as a Go method.
func (v ValidationTemplate) Render() string {
	return renderTemplate(validationTemplate, v)
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

	keys := sortedKeys(paths)
	for _, path := range keys {
		p := paths[path]
		if p.Ref != "" {
			fmt.Printf("[WARN] TODO: skipping path for %q, since it is a reference\n", path)
			continue
		}
		ops := p.Operations()
		keys := sortedKeys(ops)
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

				fields := make([]TypeField, 0)
				for _, p := range o.Parameters {
					if p.Ref != "" {
						fmt.Printf("[WARN] TODO: skipping parameter for %q, since it is a reference\n", p.Value.Name)
						continue
					}

					paramName := strcase.ToCamel(p.Value.Name)
					paramType := convertToValidGoType("", "", p.Value.Schema)
					field := TypeField{
						Name:          paramName,
						Type:          paramType,
						MarshalKey:    p.Value.Name,
						OmitDirective: "omitempty",
					}

					if p.Value.Required {
						requiredFields = requiredFields + fmt.Sprintf("\n// - %s", paramName)
					}

					fields = append(fields, field)
				}
				if o.RequestBody != nil {
					var field TypeField
					// The Nexus API spec only has a single value for content, so we can safely
					// break when a condition is met
					for mt, r := range o.RequestBody.Value.Content {
						// TODO: Handle other mime types in a more idiomatic way
						if mt != "application/json" {
							field = TypeField{
								Name:       "Body",
								Type:       "io.Reader",
								MarshalKey: "body",
								Schema:     nil, // no schema for non-JSON body
							}
							break
						}

						field = TypeField{
							Name:       "Body",
							Type:       "*" + convertToValidGoType("", "", r.Schema),
							MarshalKey: "body",
							Schema:     nil, // Body uses special serialization
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

	keys := sortedKeys(paths)
	for _, path := range keys {
		p := paths[path]
		if p.Ref != "" {
			fmt.Printf("[WARN] TODO: skipping path for %q, since it is a reference\n", path)
			continue
		}
		ops := p.Operations()
		keys := sortedKeys(ops)
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

	keys := sortedKeys(schemas)
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
	keys := sortedKeys(enumStrCollection)
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

// writeTypes iterates over the templates, constructs the different types and writes to file.
func writeTypes(f *os.File, typeCollection []TypeTemplate, typeValidationCollection []ValidationTemplate, enumCollection []EnumTemplate) {
	for _, tt := range typeCollection {
		fmt.Fprint(f, tt.Render())
	}

	for _, vm := range typeValidationCollection {
		fmt.Fprint(f, vm.Render())
	}

	for _, et := range enumCollection {
		fmt.Fprint(f, et.Render())
	}
}

// populateTypeTemplates populates the template of a type definition for the given schema.
// The enumFieldName parameter is only used as a suffix for the type name (for oneOf types).
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
	if slices.Contains(emptyTypes(), name) {
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
		properties := sortedKeys(s.Properties)
		for _, k := range properties {
			v := s.Properties[k]
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
	required := schema.Required
	fields := []TypeField{}
	keys := sortedKeys(schemas)
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

		// When `additionalProperties` is set, the type will be a map.
		// See the spec for details: https://spec.openapis.org/oas/v3.0.3.html#x4-7-24-3-3-model-with-map-dictionary-properties.
		//
		// TODO correctness: Currently our API spec does not specify
		// what type the key will be, so we set it to string to avoid
		// errors.  If the type of the key is defined in our spec in
		// the future, this should be changed to reflect that type.
		if v.Value.AdditionalProperties.Schema != nil {
			if v.Value.AdditionalProperties.Schema.Value.Type.Is("array") {
				// When `additionalProperties` has a schema of
				// type "array", use a map of string to a slice
				// of the nested type.
				typeName = fmt.Sprintf("map[string][]%s", typeName)
			} else {
				// If the schema type isn't explicitly set to
				// "array", use a map of string to the nested
				// type.
				typeName = fmt.Sprintf("map[string]%s", typeName)
			}
		}

		isRequired := slices.Contains(required, k)
		field := TypeField{
			Name:       strcase.ToCamel(k),
			Type:       typeName,
			MarshalKey: k,
			Schema:     v,
			Required:   isRequired,
		}
		// Note: pointer prefix is applied by TypeField.GoType() based on IsPointer()

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
	enumTpls := make([]EnumTemplate, 0)
	typeTpls := make([]TypeTemplate, 0)

	// First pass: find discriminator key. There must be exactly zero or one.
	discriminatorKeys := map[string]struct{}{}
	// Also find the non-discriminator property names per variant.
	valueProps := map[string]struct{}{}
	// Track whether all variants have the same single value property
	allVariantsHaveSameValueProp := true
	var commonValueProp string

	for i, variantRef := range s.OneOf {
		variantValueProps := []string{}
		for _, propName := range sortedKeys(variantRef.Value.Properties) {
			propRef := variantRef.Value.Properties[propName]

			if len(propRef.Value.Enum) == 1 {
				discriminatorKeys[propName] = struct{}{}
			} else if len(propRef.Value.Enum) > 1 {
				fmt.Printf("[WARN] TODO: oneOf for %q -> %q enum %#v\n", name, propName, propRef.Value.Enum)
			} else {
				valueProps[propName] = struct{}{}
				variantValueProps = append(variantValueProps, propName)
			}
		}
		// Check if this variant has exactly one value prop that matches others
		if len(variantValueProps) != 1 {
			allVariantsHaveSameValueProp = false
		} else if i == 0 {
			commonValueProp = variantValueProps[0]
		} else if variantValueProps[0] != commonValueProp {
			allVariantsHaveSameValueProp = false
		}
	}

	// Check invariant: there must be exactly zero or one discriminator field.
	if len(discriminatorKeys) > 1 {
		panic(fmt.Sprintf("[ERROR] Found multiple discriminator properties for type %s: %+v", name, discriminatorKeys))
	}

	// Get the discriminator (if any).
	var discriminator *Discriminator
	if len(discriminatorKeys) > 0 {
		discriminatorKeys := slices.Collect(maps.Keys(discriminatorKeys))
		discriminator = newDiscriminator(discriminatorKeys[0], typeName)
	}

	// If we have a discriminator but variants don't all share the same single value property,
	// use flat struct pattern. This includes:
	// - Variants with different value properties (e.g., AllowedSourceIps has {allow: "any"} and {allow: "list", ips: [...]})
	// - Variants with only a discriminator and no value properties (e.g., PhysicalDiskPolicy has {kind: "in_service"} and {kind: "expunged"})
	if discriminator != nil && !allVariantsHaveSameValueProp {
		return createFlatOneOf(s, name, typeName, discriminator, valueProps)
	}

	// If no value properties and no discriminator, check for other patterns
	if len(valueProps) == 0 {
		// Check if variants have enums directly (not in properties) - e.g., NameOrIdSortMode
		// These are oneOf with simple string enum variants like: oneOf: [{enum: ["a"]}, {enum: ["b"]}]
		hasDirectEnums := false
		for _, variantRef := range s.OneOf {
			if len(variantRef.Value.Enum) > 0 {
				hasDirectEnums = true
				break
			}
		}
		if hasDirectEnums {
			// Collect all enum values and create a string enum type
			allEnums := []interface{}{}
			for _, variantRef := range s.OneOf {
				allEnums = append(allEnums, variantRef.Value.Enum...)
			}
			combinedSchema := &openapi3.Schema{
				Description: s.Description,
				Enum:        allEnums,
			}
			enums, tt, et := createStringEnum(combinedSchema, collectEnumStringTypes, typeName, typeName)
			collectEnumStringTypes = enums
			typeTpls = append(typeTpls, tt...)
			enumTpls = append(enumTpls, et...)
			return typeTpls, enumTpls
		}

		// Special case: NameOrId is a union of Name and Id, both strings
		if typeName == "NameOrId" {
			typeTpl := TypeTemplate{
				Description: formatTypeDescription(typeName, s),
				Name:        typeName,
				Type:        "string",
			}
			typeTpls = append(typeTpls, typeTpl)
			return typeTpls, enumTpls
		}

		// No discriminator and no value property - fall back to any
		typeTpl := TypeTemplate{
			Description: formatTypeDescription(typeName, s),
			Name:        typeName,
			Type:        "any",
		}
		typeTpls = append(typeTpls, typeTpl)
		return typeTpls, enumTpls
	}

	// Create interface for variant types
	interfaceName := toLowerFirstLetter(typeName) + "Variant"
	markerMethod := "is" + typeName + "Variant"

	interfaceTpl := TypeTemplate{
		Description: fmt.Sprintf("// %s is an interface for %s variants.", interfaceName, typeName),
		Name:        interfaceName,
		Type:        "interface",
		OneOfMarker: markerMethod,
	}
	typeTpls = append(typeTpls, interfaceTpl)

	// Create discriminator enum type
	if discriminator != nil {
		for _, variantRef := range s.OneOf {
			if discRef, ok := variantRef.Value.Properties[discriminator.Key]; ok {
				if len(discRef.Value.Enum) == 1 {
					enums, tt, et := createStringEnum(discRef.Value, collectEnumStringTypes, discriminator.Type, discriminator.Type)
					collectEnumStringTypes = enums
					typeTpls = append(typeTpls, tt...)
					enumTpls = append(enumTpls, et...)
				}
			}
		}
	}

	// Initialize OneOfInfo for unmarshaling
	var oneOfInfo *OneOfInfo
	if discriminator != nil {
		// Use the actual property name from the schema (e.g., "value" for FieldValue, "datum" for Datum)
		valueKey := commonValueProp
		if valueKey == "" {
			valueKey = "value" // fallback
		}
		oneOfInfo = &OneOfInfo{
			Discriminator: *discriminator,
			ValueField:    strcase.ToCamel(valueKey),
			ValueKey:      valueKey,
			ValueType:     interfaceName,
			Variants:      []OneOfVariant{},
		}
	}

	// Create implementation types for each variant
	for _, variantRef := range s.OneOf {
		// Get enum value for naming this variant
		var enumValue string
		if discriminator != nil {
			if discRef, ok := variantRef.Value.Properties[discriminator.Key]; ok {
				if len(discRef.Value.Enum) == 1 {
					enumValue = strcase.ToCamel(discRef.Value.Enum[0].(string))
				}
			}
		}

		// If no discriminator, try to derive name from the variant's properties
		if enumValue == "" {
			// Use sorted property names to create a stable suffix
			propNames := sortedKeys(variantRef.Value.Properties)
			if len(propNames) > 0 {
				enumValue = strcase.ToCamel(propNames[0])
			}
		}

		implTypeName := typeName + enumValue

		// Build fields for this variant (all non-discriminator properties)
		var fields []TypeField
		for _, propName := range sortedKeys(variantRef.Value.Properties) {
			// Skip discriminator field
			if discriminator != nil && propName == discriminator.Key {
				continue
			}

			propRef := variantRef.Value.Properties[propName]
			propField := strcase.ToCamel(propName)
			propType := convertToValidGoType(propName, implTypeName, propRef)

			// If the property is a local object, create a type for it
			if isLocalObject(propRef) {
				nestedTypeName := implTypeName + propField
				tt, et := populateTypeTemplates(nestedTypeName, propRef.Value, "")
				typeTpls = append(typeTpls, tt...)
				enumTpls = append(enumTpls, et...)
				propType = nestedTypeName
			}

			// If the property is a local enum, create a type for it
			if isLocalEnum(propRef) {
				nestedTypeName := implTypeName + propField
				tt, et := populateTypeTemplates(nestedTypeName, propRef.Value, "")
				typeTpls = append(typeTpls, tt...)
				enumTpls = append(enumTpls, et...)
				propType = nestedTypeName
			}

			fields = append(fields, TypeField{
				Name:       propField,
				Type:       propType,
				MarshalKey: propName,
				Schema:     propRef,
			})
		}

		implTpl := TypeTemplate{
			Description: formatTypeDescription(implTypeName, variantRef.Value),
			Name:        implTypeName,
			Type:        "struct",
			Fields:      fields,
			OneOfMarker: markerMethod,
		}
		typeTpls = append(typeTpls, implTpl)

		if oneOfInfo != nil {
			oneOfInfo.Variants = append(oneOfInfo.Variants, OneOfVariant{
				EnumValue: oneOfInfo.Discriminator.Type + enumValue,
				ImplType:  implTypeName,
			})
		}
	}

	// Create the main wrapper struct with just the Value field
	// (discriminator is handled by MarshalJSON/UnmarshalJSON)
	// Use the actual property name from the schema (e.g., "Value" for FieldValue, "Datum" for Datum)
	valueFieldName := "Value"
	valueFieldKey := "value"
	if commonValueProp != "" {
		valueFieldName = strcase.ToCamel(commonValueProp)
		valueFieldKey = commonValueProp
	}
	mainFields := []TypeField{
		{
			Name:          valueFieldName,
			Type:          interfaceName,
			MarshalKey:    valueFieldKey,
			OmitDirective: "omitzero",
		},
	}

	typeTpl := TypeTemplate{
		Description: formatTypeDescription(typeName, s),
		Name:        typeName,
		Type:        "struct",
		Fields:      mainFields,
		OneOfInfo:   oneOfInfo,
	}
	typeTpls = append(typeTpls, typeTpl)

	return typeTpls, enumTpls
}

// createFlatOneOf creates a flat struct for oneOf types where variants have different
// property names (not a common "value" field). This matches the original code generation
// pattern where all variant fields are included in a single struct alongside the discriminator.
// Example: AllowedSourceIps has {allow: "any"} and {allow: "list", ips: [...]}
func createFlatOneOf(s *openapi3.Schema, name, typeName string, discriminator *Discriminator, valueProps map[string]struct{}) ([]TypeTemplate, []EnumTemplate) {
	enumTpls := make([]EnumTemplate, 0)
	typeTpls := make([]TypeTemplate, 0)

	// Create discriminator enum type
	for _, variantRef := range s.OneOf {
		if discRef, ok := variantRef.Value.Properties[discriminator.Key]; ok {
			if len(discRef.Value.Enum) == 1 {
				enums, tt, et := createStringEnum(discRef.Value, collectEnumStringTypes, discriminator.Type, discriminator.Type)
				collectEnumStringTypes = enums
				typeTpls = append(typeTpls, tt...)
				enumTpls = append(enumTpls, et...)
			}
		}
	}

	// Collect all fields from all variants (discriminator + all value props)
	// The discriminator field is required (no omitempty) since the API needs it to
	// identify which variant is being used.
	fields := []TypeField{
		{
			Name:       discriminator.Field,
			Type:       discriminator.Type,
			MarshalKey: discriminator.Key,
			Required:   true,
		},
	}

	// Add all value properties from all variants
	addedProps := map[string]struct{}{}
	for _, variantRef := range s.OneOf {
		for _, propName := range sortedKeys(variantRef.Value.Properties) {
			// Skip discriminator
			if propName == discriminator.Key {
				continue
			}
			// Skip if already added
			if _, ok := addedProps[propName]; ok {
				continue
			}
			addedProps[propName] = struct{}{}

			propRef := variantRef.Value.Properties[propName]
			propField := strcase.ToCamel(propName)
			propType := convertToValidGoType(propName, typeName, propRef)

			fields = append(fields, TypeField{
				Name:       propField,
				Type:       propType,
				MarshalKey: propName,
				Schema:     propRef,
			})
		}
	}

	typeTpl := TypeTemplate{
		Description: formatTypeDescription(typeName, s),
		Name:        typeName,
		Type:        "struct",
		Fields:      fields,
	}
	typeTpls = append(typeTpls, typeTpl)

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
