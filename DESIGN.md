# OneOf Type Generation Design

Handling `oneOf` types from OpenAPI is complicated in Go because the language doesn't have tagged enums (sum types). However, we want to use concrete types wherever possible rather than falling back to `any`, which loses type safety and discoverability.

This document describes our approach to representing `oneOf` values idiomatically in Go. We detect different patterns in the OpenAPI schema and generate appropriate Go types for each.

## Pattern 1: Discriminator + Single Common Value Property (Interface Pattern)

**Detection:** All variants have a discriminator property (single enum value) and share the same single value property name.

**Generated Go:** Interface type with variant struct implementations, custom `MarshalJSON`/`UnmarshalJSON`.

**OpenAPI Example - `FieldValue`:**
```json
"FieldValue": {
  "oneOf": [
    {
      "properties": {
        "type": { "enum": ["string"] },
        "value": { "type": "string" }
      }
    },
    {
      "properties": {
        "type": { "enum": ["i8"] },
        "value": { "type": "integer", "format": "int8" }
      }
    }
  ]
}
```

**Generated Go:**
```go
type fieldValueVariant interface {
	isFieldValueVariant()
}

type FieldValueType string

type FieldValueString struct {
	Value string `json:"value,omitempty"`
}
func (FieldValueString) isFieldValueVariant() {}

type FieldValueI8 struct {
	Value *int `json:"value,omitempty"`
}
func (FieldValueI8) isFieldValueVariant() {}

type FieldValue struct {
	Value fieldValueVariant `json:"value,omitzero"`
}
func (v *FieldValue) UnmarshalJSON(data []byte) error { ... }
func (v FieldValue) MarshalJSON() ([]byte, error) { ... }
```

**OpenAPI Example - `Datum`:** (same pattern, value field named `datum`)
```json
"Datum": {
  "oneOf": [
    {
      "properties": {
        "type": { "enum": ["bool"] },
        "datum": { "type": "boolean" }
      }
    }
  ]
}
```

**Generated Go:**
```go
type datumVariant interface {
	isDatumVariant()
}

type DatumType string

type DatumBool struct {
	Datum *bool `json:"datum,omitempty"`
}
func (DatumBool) isDatumVariant() {}

type Datum struct {
	Datum datumVariant `json:"datum,omitzero"`
}
func (v *Datum) UnmarshalJSON(data []byte) error { ... }
func (v Datum) MarshalJSON() ([]byte, error) { ... }
```

---

## Pattern 2: Discriminator + Different/Varying Value Properties (Flat Struct Pattern)

**Detection:** Variants have a discriminator property but don't all share the same single value property (some have different properties, some have none).

**Generated Go:** Single flat struct with discriminator field plus all possible value fields from all variants.

**OpenAPI Example - `AllowedSourceIps`:**
```json
"AllowedSourceIps": {
  "oneOf": [
    {
      "properties": {
        "allow": { "enum": ["any"] }
      }
    },
    {
      "properties": {
        "allow": { "enum": ["list"] },
        "ips": { "type": "array", "items": { "$ref": "#/components/schemas/IpNet" } }
      }
    }
  ]
}
```

- Discriminator: `allow`
- Variant `"any"`: no value properties
- Variant `"list"`: has `ips` property

**Generated Go:**
```go
type AllowedSourceIpsAllow string

type AllowedSourceIps struct {
	Allow AllowedSourceIpsAllow `json:"allow,omitempty"`
	Ips   []IpNet               `json:"ips,omitempty"`
}
```

**OpenAPI Example - `PhysicalDiskPolicy`:** (discriminator only, no value properties in any variant)
```json
"PhysicalDiskPolicy": {
  "oneOf": [
    {
      "properties": {
        "kind": { "enum": ["in_service"] }
      }
    },
    {
      "properties": {
        "kind": { "enum": ["expunged"] }
      }
    }
  ]
}
```

- Discriminator: `kind`
- No value properties in any variant

**Generated Go:**
```go
type PhysicalDiskPolicyKind string

type PhysicalDiskPolicy struct {
	Kind PhysicalDiskPolicyKind `json:"kind,omitempty"`
}
```

---

## Pattern 3: Simple String Enum (Direct Enum Variants)

**Detection:** Variants are directly string enum values (not wrapped in an object with properties).

**Generated Go:** Simple `type X string` with const values.

**OpenAPI Example - `NameOrIdSortMode`:**
```json
"NameOrIdSortMode": {
  "oneOf": [
    { "type": "string", "enum": ["name_ascending"] },
    { "type": "string", "enum": ["name_descending"] },
    { "type": "string", "enum": ["id_ascending"] }
  ]
}
```

**Generated Go:**
```go
type NameOrIdSortMode string

const NameOrIdSortModeNameAscending NameOrIdSortMode = "name_ascending"
const NameOrIdSortModeNameDescending NameOrIdSortMode = "name_descending"
const NameOrIdSortModeIdAscending NameOrIdSortMode = "id_ascending"

var NameOrIdSortModeCollection = []NameOrIdSortMode{
	NameOrIdSortModeIdAscending,
	NameOrIdSortModeNameAscending,
	NameOrIdSortModeNameDescending,
}
```

---

## Pattern 4: Untagged Union (No Discriminator)

**Detection:** Variants use `allOf` with `$ref` to reference other types, no discriminator property.

**Generated Go:** `type X any` (cannot be automatically disambiguated without runtime type inspection).

**OpenAPI Example - `IpNet`:**
```json
"IpNet": {
  "oneOf": [
    { "title": "v4", "allOf": [{ "$ref": "#/components/schemas/Ipv4Net" }] },
    { "title": "v6", "allOf": [{ "$ref": "#/components/schemas/Ipv6Net" }] }
  ]
}
```

**Generated Go:**
```go
type IpNet any
```

**OpenAPI Example - `IpRange`:**
```json
"IpRange": {
  "oneOf": [
    { "title": "v4", "allOf": [{ "$ref": "#/components/schemas/Ipv4Range" }] },
    { "title": "v6", "allOf": [{ "$ref": "#/components/schemas/Ipv6Range" }] }
  ]
}
```

**Generated Go:**
```go
type IpRange any
```
