# Design Notes

## oneOf Type Generation

### Overview

OpenAPI `oneOf` types represent tagged unions, which are values that can be one of several variant types. Go doesn't have a native union type, so we have to decide how to represent these values in a safe and ergonomic way. There are a few distinct types of `oneOf` types in the Oxide API, related to different `serde` tagging strategies, and we handle each of them differently.

### Discriminator with single value type

When a `oneOf` has:
1. Exactly one discriminator property (a field with a single enum value per variant)
2. Exactly one multi-type property (a field whose type varies across variants)

We generate an **interface with variant wrapper types** pattern.

**Example: `FieldValue`**

The OpenAPI spec defines `FieldValue` as (abbreviated):
```yaml
FieldValue:
  oneOf:
    - type: object
      properties:
        type: { enum: ["string"] }
        value: { type: string }
    - type: object
      properties:
        type: { enum: ["i8"] }
        value: { type: integer, format: int8 }
    - type: object
      properties:
        type: { enum: ["bool"] }
        value: { type: boolean }
    # ... more variants for u8, i16, u16, i32, u32, i64, u64, ip_addr, uuid
```

This has:
- One discriminator: `type` (with values `string`, `i8`, `bool`, etc.)
- One multi-type property: `value` (which is `string`, `int`, or `bool` depending on variant)

So it generates:

```go
// Interface that all variants implement
type fieldValueVariant interface {
    isFieldValueVariant()
}

// Variant wrapper types (one per discriminator value)
type FieldValueString struct {
    Value string `json:"value,omitzero"`
}
func (FieldValueString) isFieldValueVariant() {}

type FieldValueI8 struct {
    Value *int `json:"value,omitzero"`
}
func (FieldValueI8) isFieldValueVariant() {}

type FieldValueBool struct {
    Value *bool `json:"value,omitzero"`
}
func (FieldValueBool) isFieldValueVariant() {}

// Main type with only the value field
type FieldValue struct {
    Value fieldValueVariant `json:"value,omitzero"`
}

// Type() method derives the discriminator from the concrete Value type
func (v FieldValue) Type() FieldValueType {
    switch v.Value.(type) {
    case *FieldValueString:
        return FieldValueTypeString
    case *FieldValueI8:
        return FieldValueTypeI8
    // ... etc
    }
}
```

The discriminator field is not stored in the struct. We don't want it to be a
public member of the struct, because users would then be able to set its value
and cause it to mismatch the type of the value field. Instead, we expose a
public method named after the discriminator that returns the discriminator value.

We also implement custom `MarshalJSON` and `UnmarshalJSON` methods for the main
type. To unmarshal, we check the discriminator field in the JSON to determine
which concrete type to use for unmarshalling the value. To marshal, we call the
`Type()` method to determine which discriminator to emit.

### Discriminator without a single value type

When a `oneOf` doesn't meet the interface pattern criteria, we use a flat struct
that contains all properties from all variants. Properties that have different
types across variants become `any`.

**Example: `DiskSource`**

The OpenAPI spec defines `DiskSource` as:
```yaml
DiskSource:
  oneOf:
    - type: object
      properties:
        type: { enum: ["blank"] }
        block_size: { $ref: "#/components/schemas/BlockSize" }
    - type: object
      properties:
        type: { enum: ["snapshot"] }
        snapshot_id: { type: string, format: uuid }
    - type: object
      properties:
        type: { enum: ["image"] }
        image_id: { type: string, format: uuid }
    - type: object
      properties:
        type: { enum: ["importing_blocks"] }
        block_size: { $ref: "#/components/schemas/BlockSize" }
```

This has a discriminator (`type`) but no multi-type property. Each variant has
different fields (`block_size`, `snapshot_id`, `image_id`), not different types
for the same field. So we generate a flat struct:

```go
type DiskSource struct {
    BlockSize  BlockSize      `json:"block_size,omitempty"`
    Type       DiskSourceType `json:"type,omitempty"`
    SnapshotId string         `json:"snapshot_id,omitempty"`
    ImageId    string         `json:"image_id,omitempty"`
}
```

If any property had different types across variants, it would become `any`.

### Untagged union

When a `oneOf` has no object properties (i.e., variants are primitive types or
references wrapped in `allOf`), the type becomes `interface{}`.

**Example: `IpNet`**

The OpenAPI spec defines `IpNet` as:
```yaml
IpNet:
  oneOf:
    - title: v4
      allOf:
        - $ref: "#/components/schemas/Ipv4Net"
    - title: v6
      allOf:
        - $ref: "#/components/schemas/Ipv6Net"
```

```go
type IpNet interface{}
```

Note: we may be able to handle these types better in the future. For example, we
could detect that all variants are effectively strings and represent `IpNet` as
`string`. Alternatively, we could represent `Ipv4Net` and `Ipv6Net` as distinct
types with their own validation logic, and attempt to unmarshal into each
variant type until we find a match.
