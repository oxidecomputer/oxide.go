# Design Notes

## Goals and non-goals

The code generation logic in internal/ is developed and tested against the Nexus API and its OpenAPI
specification. Nexus is built with Rust and uses crates like serde, schemars, and dropshot to build
its OpenAPI spec, so we focus on the patterns used by those tools. We don't aim to support all
OpenAPI features or patterns as of this writing.

For example, Nexus represents tagged unions using inline object definitions, but represents untagged
unions with references to schemas defined elsewhere in the spec. Different tools might generate
OpenAPI spec files where tagged unions are instead represented as schema references, or untagged
unions as inline objects, but we don't support those cases because they can't occur in Nexus.

## oneOf Type Generation

### Overview

OpenAPI `oneOf` types represent sum types, which are values that can be one of several variant
types. Go doesn't have a native sum type, so we have to decide how to represent these values in a
safe and ergonomic way. There are a few distinct patterns of `oneOf` types in the Oxide API, related
to different `serde` tagging strategies, and we handle each of them differently.

### Tagged union

When a `oneOf` has:

1. Exactly one discriminator property (a field with a single enum value per variant)
2. Exactly one multi-type property (a field whose type varies across variants)

We generate an **interface with variant wrapper types** pattern.

**Example: `PrivateIpStack`**

In Rust, `PrivateIpStack` is defined as:

```rust
#[serde(tag = "type", content = "value", rename_all = "snake_case")]
pub enum PrivateIpStack {
    V4(PrivateIpv4Stack),
    V6(PrivateIpv6Stack),
    DualStack { v4: PrivateIpv4Stack, v6: PrivateIpv6Stack },
}
```

This generates the following OpenAPI spec:

```yaml
PrivateIpStack:
  oneOf:
    - type: object
      properties:
        type: { enum: ["v4"] }
        value: { $ref: "#/components/schemas/PrivateIpv4Stack" }
      required: [type, value]
    - type: object
      properties:
        type: { enum: ["v6"] }
        value: { $ref: "#/components/schemas/PrivateIpv6Stack" }
      required: [type, value]
    - type: object
      properties:
        type: { enum: ["dual_stack"] }
        value:
          type: object
          properties:
            v4: { $ref: "#/components/schemas/PrivateIpv4Stack" }
            v6: { $ref: "#/components/schemas/PrivateIpv6Stack" }
      required: [type, value]
```

This has:

- One discriminator: `type` (with values `v4`, `v6`, `dual_stack`)
- One multi-type property: `value` (which is `PrivateIpv4Stack`, `PrivateIpv6Stack`, or an inline
  object depending on variant)

So it generates:

```go
// Interface with marker method that all variants implement
type privateIpStackVariant interface {
    isPrivateIpStackVariant()
}

// Variant wrapper types (one per discriminator value)
type PrivateIpStackV4 struct {
    Value PrivateIpv4Stack `json:"value"`
}
func (PrivateIpStackV4) isPrivateIpStackVariant() {}

type PrivateIpStackV6 struct {
    Value PrivateIpv6Stack `json:"value"`
}
func (PrivateIpStackV6) isPrivateIpStackVariant() {}

type PrivateIpStackDualStack struct {
    Value PrivateIpStackValue `json:"value"`
}
func (PrivateIpStackDualStack) isPrivateIpStackVariant() {}

// Main type with only the value field
type PrivateIpStack struct {
    Value privateIpStackVariant `json:"value,omitempty"`
}

// Type() method derives the discriminator from the concrete Value type
func (v PrivateIpStack) Type() PrivateIpStackType {
    switch v.Value.(type) {
    case *PrivateIpStackV4:
        return PrivateIpStackTypeV4
    case *PrivateIpStackV6:
        return PrivateIpStackTypeV6
    case *PrivateIpStackDualStack:
        return PrivateIpStackTypeDualStack
    default:
        return ""
    }
}
```

The discriminator field is not stored in the struct. We don't want users to be able to set its value
so that it doesn't match the type of the value field. Instead, we expose a public method named after
the discriminator that returns the discriminator value.

We also implement custom `MarshalJSON` and `UnmarshalJSON` methods for the main type. To unmarshal,
we check the discriminator field in the JSON to determine which concrete type to use for
unmarshalling the value. To marshal, we call the `Type()` method to determine which discriminator to
emit.

**Usage examples:**

```go
// Reading a network interface's IP stack from the API
nic, _ := client.InstanceNetworkInterfaceView(ctx, params)
ipStack := nic.IpStack

switch v := ipStack.Value.(type) {
case *oxide.PrivateIpStackV4:
    fmt.Printf("IPv4 only: %s\n", v.Value.Ip)
case *oxide.PrivateIpStackV6:
    fmt.Printf("IPv6 only: %s\n", v.Value.Ip)
case *oxide.PrivateIpStackDualStack:
    fmt.Printf("Dual stack: %s / %s\n", v.Value.V4.Ip, v.Value.V6.Ip)
}
```

```go
// Creating a network interface with an IPv4-only stack
params := oxide.InstanceNetworkInterfaceCreateParams{
    Body: &oxide.InstanceNetworkInterfaceCreate{
        Name:       "my-nic",
        SubnetName: "my-subnet",
        VpcName:    "my-vpc",
        IpConfig: oxide.PrivateIpStackCreate{
            Value: &oxide.PrivateIpStackCreateV4{
                Value: oxide.PrivateIpv4StackCreate{
                    Ip: oxide.Ipv4Assignment{
                        Type:  oxide.Ipv4AssignmentTypeExplicit,
                        Value: "10.0.0.5",
                    },
                },
            },
        },
    },
}
```

**Why wrapper structs?**

We represent each variant type using a wrapper struct, e.g.

```go
type PrivateIpStackV4 struct {
    Value PrivateIpv4Stack `json:"value"`
}
```

The wrapper struct implements the marker method for the relevant interface:

```go
func (PrivateIpStackV4) isPrivateIpStackVariant() {}
```

Why use wrapper structs? For some variant types, we could omit the wrapper and implement the marker
method on the wrapped type instead:

```go
func (PrivateIpv4Stack) isPrivateIpStackVariant() {}
```

Primitive types can't implement methods, but we could use type definitions instead:

```go
type MyPrimitiveVariant string

func (MyPrimitiveVariant) isMyTypeVariant() {}
```

However, this presents a few problems:

- Some variant types are `interface{}` or `any`. We can't implement methods on `any`.
- Some variant types are represented by pointers to primitive types, like `*bool` or `*int`. We
  can't implement methods on pointers to primitive types.

We could represent some variant types with wrapper structs and others as unwrapped structs or type
definitions. But the complexity of conditionally wrapping variant types is potentially more
confusing to end users than consistent use of wrappers.

Note: we can reconsider this choice if we're able to drop the use of `interface{}` types and
pointers to primitives for variants, and if we're confident that those cases won't emerge again.

### Discriminator with multiple value fields

When a `oneOf` has a discriminator field and _multiple_ value fields, we use a flat struct that
contains all properties from all variants. Properties that have different types across variants
become `any`.

**Example: `DiskSource`**

In Rust, `DiskSource` is defined as:

```rust
#[serde(tag = "type", rename_all = "snake_case")]
pub enum DiskSource {
    Blank { block_size: BlockSize },
    Snapshot { snapshot_id: Uuid },
    Image { image_id: Uuid },
    ImportingBlocks { block_size: BlockSize },
}
```

This generates the following OpenAPI spec:

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

This has a discriminator (`type`) but no multi-type property. Each variant has different fields
(`block_size`, `snapshot_id`, `image_id`), not different types for the same field. So we generate a
flat struct:

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

When a `oneOf` schema has no discriminator property (i.e., it's defined in Rust using
`serde(untagged)`), we can't use the discriminator to determine the correct variant type for
unmarshalling. Instead, if the variants use OpenAPI `format` or `pattern` fields, we use those to
choose the variant type. In this case, we use the interface with marker methods pattern, as for
tagged unions.

Untagged unions are detected when:

1. Each variant is an `allOf` wrapper containing a single `$ref`
2. The referenced types can be discriminated by either:
   - **Format-based**: Object types where fields have distinct `format` values (e.g., `ipv4` vs
     `ipv6`)
   - **Pattern-based**: String types with distinct regex `pattern` values

**Example: `IpNet` (pattern-based)**

In Rust, `IpNet` is defined as:

```rust
#[serde(untagged)]
pub enum IpNet {
    V4(Ipv4Net),
    V6(Ipv6Net),
}
```

This generates the following OpenAPI spec:

```yaml
IpNet:
  oneOf:
    - title: v4
      allOf:
        - $ref: "#/components/schemas/Ipv4Net"
    - title: v6
      allOf:
        - $ref: "#/components/schemas/Ipv6Net"

Ipv4Net:
  type: string
  pattern: "^([0-9]{1,3}\\.){3}[0-9]{1,3}/[0-9]{1,2}$"

Ipv6Net:
  type: string
  pattern: "^[0-9a-fA-F:]+/[0-9]{1,3}$"
```

Since `Ipv4Net` and `Ipv6Net` are string types with distinct regex patterns, we generate:

```go
// Interface with marker method
type ipNetVariant interface {
    isIpNetVariant()
}

// Marker methods on existing types
func (Ipv4Net) isIpNetVariant() {}
func (Ipv6Net) isIpNetVariant() {}

// Wrapper struct
type IpNet struct {
    Value ipNetVariant `json:"value,omitempty"`
}

// Pattern-based discrimination using compiled regexes
var (
    ipv4netPattern = regexp.MustCompile(`^...`)
    ipv6netPattern = regexp.MustCompile(`^...`)
)

func (v *IpNet) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    if ipv4netPattern.MatchString(s) {
        val := Ipv4Net(s)
        v.Value = &val
        return nil
    }
    if ipv6netPattern.MatchString(s) {
        val := Ipv6Net(s)
        v.Value = &val
        return nil
    }
    return fmt.Errorf("no variant matched for IpNet: %q", s)
}
```

**Example: `IpRange` (format-based)**

In Rust, `IpRange` is defined as:

```rust
#[serde(untagged)]
pub enum IpRange {
    V4(Ipv4Range),
    V6(Ipv6Range),
}
```

This generates the following OpenAPI spec:

```yaml
IpRange:
  oneOf:
    - title: v4
      allOf:
        - $ref: "#/components/schemas/Ipv4Range"
    - title: v6
      allOf:
        - $ref: "#/components/schemas/Ipv6Range"

Ipv4Range:
  type: object
  properties:
    first: { type: string, format: ipv4 }
    last: { type: string, format: ipv4 }

Ipv6Range:
  type: object
  properties:
    first: { type: string, format: ipv6 }
    last: { type: string, format: ipv6 }
```

Since `Ipv4Range` and `Ipv6Range` have fields with distinct `format` values, we generate:

```go
// Interface with marker method
type ipRangeVariant interface {
    isIpRangeVariant()
}

// Marker methods on existing types
func (Ipv4Range) isIpRangeVariant() {}
func (Ipv6Range) isIpRangeVariant() {}

// Wrapper struct
type IpRange struct {
    Value ipRangeVariant `json:"value,omitempty"`
}

// Format detection functions (call DetectXxxFormat from format_detectors.go)
func detectIpv4Range(v *Ipv4Range) bool {
    if !DetectIpv4Format(v.First) {
        return false
    }
    if !DetectIpv4Format(v.Last) {
        return false
    }
    return true
}

func detectIpv6Range(v *Ipv6Range) bool {
    if !DetectIpv6Format(v.First) {
        return false
    }
    if !DetectIpv6Format(v.Last) {
        return false
    }
    return true
}

func (v *IpRange) UnmarshalJSON(data []byte) error {
    // Try Ipv4Range
    {
        var candidate Ipv4Range
        if err := json.Unmarshal(data, &candidate); err == nil {
            if detectIpv4Range(&candidate) {
                v.Value = &candidate
                return nil
            }
        }
    }
    // Try Ipv6Range
    {
        var candidate Ipv6Range
        if err := json.Unmarshal(data, &candidate); err == nil {
            if detectIpv6Range(&candidate) {
                v.Value = &candidate
                return nil
            }
        }
    }
    return fmt.Errorf("no variant matched for IpRange: %s", string(data))
}
```

Note that we only use the `format` and `pattern` fields for variant type detection, not for
validation. In the future, we may consider validating based on `format` and/or `pattern` during
unmarshalling, marshalling, or both. For now, we trust the API to send valid data and error when
receiving bad data.

**Usage examples:**

```go
// Reading an IP range from the API
poolRange, _ := client.IpPoolRangeList(ctx, params)
for _, item := range poolRange.Items {
    switch v := item.Range.Value.(type) {
    case *oxide.Ipv4Range:
        fmt.Printf("IPv4: %s - %s\n", v.First, v.Last)
    case *oxide.Ipv6Range:
        fmt.Printf("IPv6: %s - %s\n", v.First, v.Last)
    }
}
```

```go
// Creating an IP range
ipRange := oxide.IpRange{Value: &oxide.Ipv4Range{
    First: "192.168.1.1",
    Last:  "192.168.1.100",
}}
```

**Fallback behavior:**

If we cannot distinguish between variant types, we fall back to generating `interface{}`. This
happens when:

- Variants are not wrapped in `allOf` with a single `$ref`
- Not all variants have regex patterns (for pattern-based discrimination)
- Not all variants have format-constrained fields (for format-based discrimination)
