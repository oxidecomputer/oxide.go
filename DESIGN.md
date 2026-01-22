# Design Notes

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

When a `oneOf` has no object properties (i.e., variants are primitive types or references wrapped in
`allOf`), the type becomes `interface{}`.

**Example: `IpNet`**

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
```

```go
type IpNet interface{}
```

Note: we may be able to handle these types better in the future. For example, we could detect that
all variants are effectively strings and represent `IpNet` as `string`. Alternatively, we could
represent `Ipv4Net` and `Ipv6Net` as distinct types with their own validation logic, and attempt to
unmarshal into each variant type until we find a match.
