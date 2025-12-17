// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"bytes"
	"fmt"
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
			Funcs(template.FuncMap{"splitDocString": splitDocString}).
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
	// Type describes the type of the type (e.g. struct, int64, string)
	Type string
	// Fields holds the information for the field
	Fields []TypeField
}

// Render renders the TypeTemplate to a Go type.
func (t TypeTemplate) Render() string {
	return renderTemplate(typeTemplate, t)
}

// ToValidation converts a TypeTemplate to a ValidationTemplate.
// Returns nil if the type is not a struct.
// Always generates a Validate method for struct types (even if empty) so that
// nested types can consistently call Validate().
func (t TypeTemplate) ToValidation() *ValidationTemplate {
	if t.Type != "struct" || t.Name == "" {
		return nil
	}

	fields := []FieldValidation{}
	for _, f := range t.Fields {
		if fv := fieldValidationFromTypeField(f); fv != nil {
			fields = append(fields, *fv)
		}
	}

	// Always return a ValidationTemplate for struct types, even with no fields
	return &ValidationTemplate{
		AssociatedType: t.Name,
		Fields:         fields,
	}
}

// fieldValidationFromTypeField creates a FieldValidation from a TypeField.
func fieldValidationFromTypeField(f TypeField) *FieldValidation {
	if f.Schema == nil {
		return nil
	}

	// Skip validation for 'any' type fields - can't validate interface{}
	if f.Type == "any" {
		if f.Required {
			return &FieldValidation{
				Name:     f.Name,
				JSONName: f.MarshalKey,
				Required: true,
				Type:     "object",
			}
		}
		return nil
	}

	return buildFieldValidation(f.Name, f.MarshalKey, f.Schema, f.Required)
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
	case f.Schema == nil:
		omitDirective = "omitempty"
	case f.Required || isNullableArray(f.Schema):
		omitDirective = ""
	case slices.Contains(omitzeroTypes(), f.Type):
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
	AssociatedType string
	Fields         []FieldValidation // slice for deterministic code generation
}

// FieldValidation holds all validation rules for a single field.
type FieldValidation struct {
	Name           string // Go field name
	JSONName       string // JSON field name for error messages
	Required       bool
	Type           string // "string", "int", "object" - determines required check variant
	Pattern        string // regex pattern
	Format         string // uuid, email, ipv4, ipv6, uri, hostname, date-time
	EnumType       string // enum type name (e.g., "AddressLotKind")
	CollectionName string // enum collection var name (e.g., "AddressLotKindCollection")
	IsNested       bool   // field type has a Validate() method
	IsSlice        bool   // field is a slice - iterate and validate each element
	IsPointer      bool   // field is a pointer - nil check before validation
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
					Fields:         []FieldValidation{},
				}

				for _, p := range o.Parameters {
					fieldVal := buildFieldValidation(
						strcase.ToCamel(p.Value.Name),
						p.Value.Name,
						p.Value.Schema,
						p.Value.Required,
					)
					if fieldVal != nil {
						validationTpl.Fields = append(validationTpl.Fields, *fieldVal)
					}
				}

				if o.RequestBody != nil {
					// If an endpoint has a body, our API requires it, so we can safely add it.
					bodyField := FieldValidation{
						Name:     "Body",
						JSONName: "body",
						Required: true,
						Type:     "object",
					}

					// Check if the body is JSON and is a struct type (not interface)
					// Only struct types have Validate methods
					for mt, mediaType := range o.RequestBody.Value.Content {
						if mt == "application/json" && mediaType.Schema != nil {
							schema := mediaType.Schema.Value
							// Don't mark as nested if it's an interface type (oneOf/anyOf)
							// or if it has no properties (not a struct)
							isInterface := len(schema.OneOf) > 0 || len(schema.AnyOf) > 0
							isStruct := schema.Type.Is("object") && len(schema.Properties) > 0
							// Also check referenced types
							isRefToStruct := mediaType.Schema.Ref != "" && len(schema.Properties) > 0

							if !isInterface && (isStruct || isRefToStruct) {
								bodyField.IsNested = true
								bodyField.IsPointer = true // Body is always a pointer type for JSON
							}
						}
						break // Only one content type per endpoint
					}

					validationTpl.Fields = append(validationTpl.Fields, bodyField)
				}
				validationMethods = append(validationMethods, validationTpl)
			}
		}

	}

	return validationMethods
}

// buildFieldValidation constructs a FieldValidation from an OpenAPI schema.
func buildFieldValidation(goName, jsonName string, schemaRef *openapi3.SchemaRef, required bool) *FieldValidation {
	if schemaRef == nil || schemaRef.Value == nil {
		return nil
	}

	schema := schemaRef.Value

	// Skip fields with no defined type (becomes 'any' in Go) - can't validate interface{}
	if schema.Type == nil || len(schema.Type.Slice()) == 0 {
		// Only return a basic required check for 'any' fields
		if required {
			return &FieldValidation{
				Name:     goName,
				JSONName: jsonName,
				Required: true,
				Type:     "object", // Use object for nil check
			}
		}
		return nil
	}

	fieldVal := &FieldValidation{
		Name:     goName,
		JSONName: jsonName,
		Required: required,
	}

	// Check for interface types (oneOf/anyOf) - these don't have Validate methods
	isInterfaceType := len(schema.OneOf) > 0 || len(schema.AnyOf) > 0

	// Check for map types (additionalProperties) - these don't have Validate methods
	isMapType := schema.AdditionalProperties.Schema != nil

	// Determine field type for required checks
	// Note: date-time/date/time formats become *time.Time in Go, so treat as object
	isTimeFormat := schema.Format == "date-time" || schema.Format == "date" || schema.Format == "time"

	// Integer formats that don't become *int in Go (uint64, uint32, etc.)
	// These become type aliases and should be treated as object for required checks
	isSpecialIntFormat := schema.Format == "uint64" || schema.Format == "uint32" ||
		schema.Format == "uint16" || schema.Format == "uint8" ||
		schema.Format == "int64" || schema.Format == "int32" ||
		schema.Format == "int16" || schema.Format == "int8"

	if schema.Type.Is("string") && !isTimeFormat {
		fieldVal.Type = "string"
	} else if schema.Type.Is("integer") && !isSpecialIntFormat {
		fieldVal.Type = "int"
	} else if schema.Type.Is("number") {
		// float64 in Go - treat as object for required checks (no HasRequiredNum for float64)
		fieldVal.Type = "object"
	} else {
		fieldVal.Type = "object"
	}

	// Only apply pattern/format validation to actual string types (not time types)
	if schema.Type.Is("string") && !isTimeFormat {
		// Extract pattern
		if schema.Pattern != "" {
			fieldVal.Pattern = schema.Pattern
		}

		// Extract format
		if schema.Format != "" {
			fieldVal.Format = schema.Format
		}
	}

	// Check for enum - only string enums have collections generated
	// Integer enums (like BlockSize) don't have collections, so skip them
	if schemaRef.Ref != "" && len(schema.Enum) > 0 && schema.Type.Is("string") {
		enumTypeName := getReferenceSchema(schemaRef)
		fieldVal.EnumType = enumTypeName
		fieldVal.CollectionName = enumTypeName + "Collection"
	}

	// Check if this field needs nested validation (only actual struct types)
	// Don't mark as nested if:
	// - It's an interface type (oneOf/anyOf)
	// - It's a map type (additionalProperties)
	// - It's a string/enum type
	// - It has no properties (not a struct)
	if !isInterfaceType && !isMapType {
		isStructType := schema.Type.Is("object") && len(schema.Properties) > 0
		isRefToStruct := schemaRef.Ref != "" && !schema.Type.Is("string") && len(schema.Enum) == 0 && len(schema.Properties) > 0
		if isStructType || isRefToStruct {
			fieldVal.IsNested = true
		}
	}

	// Check for array types
	if schema.Type.Is("array") {
		fieldVal.IsSlice = true
		// Only mark as nested if array items are struct types (not interfaces, not strings)
		if schema.Items != nil && schema.Items.Value != nil {
			itemSchema := schema.Items.Value
			isItemInterface := len(itemSchema.OneOf) > 0 || len(itemSchema.AnyOf) > 0
			isItemStruct := itemSchema.Type.Is("object") && len(itemSchema.Properties) > 0
			isItemRefToStruct := schema.Items.Ref != "" && !itemSchema.Type.Is("string") && len(itemSchema.Enum) == 0 && len(itemSchema.Properties) > 0
			if !isItemInterface && (isItemStruct || isItemRefToStruct) {
				fieldVal.IsNested = true
			}
		}
	}

	// Check if pointer type
	if schema.Nullable {
		fieldVal.IsPointer = true
	}

	return fieldVal
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

	// Build a set of types that already have validation from Params
	hasValidation := make(map[string]bool)
	for _, vm := range typeValidationCollection {
		hasValidation[vm.AssociatedType] = true
	}

	// Generate validation for Params types
	for _, vm := range typeValidationCollection {
		fmt.Fprint(f, vm.Render())
	}

	// Generate validation for schema types (structs) that don't already have validation
	for _, tt := range typeCollection {
		if hasValidation[tt.Name] {
			continue // Skip types that already have Params validation
		}
		if v := tt.ToValidation(); v != nil {
			fmt.Fprint(f, v.Render())
		}
	}

	for _, et := range enumCollection {
		fmt.Fprint(f, et.Render())
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
	enumTpls := make([]EnumTemplate, 0)
	typeTpls := make([]TypeTemplate, 0)

	// Loop over variants, creating types and enums for nested types, and gathering metadata about the oneOf overall.

	// Set of candidate discriminator keys. There must be exactly zero or one discriminator key.
	discriminatorKeys := map[string]struct{}{}
	// Map of properties to sets of variant types. We use this to identify fields with multiple types across variants.
	propToVariantTypes := map[string]map[string]struct{}{}

	for _, variantRef := range s.OneOf {
		enumField := ""
		for _, propName := range sortedKeys(variantRef.Value.Properties) {
			propRef := variantRef.Value.Properties[propName]
			propField := strcase.ToCamel(propName)

			if len(propRef.Value.Enum) == 1 {
				discriminatorKeys[propName] = struct{}{}
				enumField = strcase.ToCamel(propRef.Value.Enum[0].(string))
			} else if len(propRef.Value.Enum) > 1 {
				fmt.Printf("[WARN] TODO: oneOf for %q -> %q enum %#v\n", name, propName, propRef.Value.Enum)
			} else if propRef.Value.Enum == nil && len(variantRef.Value.Properties) == 1 {
				enumField = propField
			}
			if _, ok := propToVariantTypes[propName]; !ok {
				propToVariantTypes[propName] = map[string]struct{}{}
			}
			goType := convertToValidGoType(propName, typeName, propRef)
			propToVariantTypes[propName][goType] = struct{}{}
		}
		tt, et := populateTypeTemplates(name, variantRef.Value, enumField)
		typeTpls = append(typeTpls, tt...)
		enumTpls = append(enumTpls, et...)
	}

	// Check invariant: there must be exactly zero or one discriminator field.
	if len(discriminatorKeys) > 1 {
		panic(fmt.Sprintf("[ERROR] Found multiple discriminator properties for type %s: %+v", name, discriminatorKeys))
	}

	// Find properties that have different types across variants.
	multiTypeProps := map[string]struct{}{}
	for propName, variantTypes := range propToVariantTypes {
		if len(variantTypes) > 1 {
			multiTypeProps[propName] = struct{}{}
		}
	}

	// Build the struct type for the oneOf field, if defined.
	oneOfFields := []TypeField{}
	seenFields := map[string]struct{}{}
	for _, variantRef := range s.OneOf {
		for _, propName := range sortedKeys(variantRef.Value.Properties) {
			if _, ok := seenFields[propName]; ok {
				continue
			}
			seenFields[propName] = struct{}{}

			propRef := variantRef.Value.Properties[propName]
			propField := strcase.ToCamel(propName)
			propType := convertToValidGoType(propName, typeName, propRef)

			// Use the enum type name instead of "string" when the property has an enum.
			if propType == "string" && len(propRef.Value.Enum) != 0 {
				propType = typeName + strcase.ToCamel(propName)
			}

			// Use "any" if this property has different types across variants.
			if _, ok := multiTypeProps[propName]; ok {
				propType = "any"
			}

			// Determine omit directive: nullable fields in oneOf use omitzero.
			var omitDirective string
			if propRef.Value != nil && propRef.Value.Nullable {
				omitDirective = "omitzero"
			}

			field := TypeField{
				Name:                propField,
				Type:                propType,
				MarshalKey:          propName,
				Schema:              propRef,
				FallbackDescription: true,
				OmitDirective:       omitDirective,
			}

			oneOfFields = append(oneOfFields, field)
		}
	}

	if len(oneOfFields) > 0 {
		typeTpl := TypeTemplate{
			Description: formatTypeDescription(typeName, s),
			Name:        typeName,
			Type:        "struct",
			Fields:      oneOfFields,
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
