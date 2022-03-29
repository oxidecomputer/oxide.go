package oxide

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Equal returns true if `other` is a ByteCount and has the same value as `i`.
func (i ByteCount) Equal(other attr.Value) bool {
	o, ok := other.(ByteCount)

	if !ok {
		return false
	}

	return i == o
}

// ToTerraformValue returns the data contained in the ByteCount as a tftypes.Value.
func (i ByteCount) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {

	if err := tftypes.ValidateValue(tftypes.Number, i); err != nil {
		return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), err
	}
	return tftypes.NewValue(tftypes.Number, i), nil
}

// Type returns a NumberType.
func (i ByteCount) Type(ctx context.Context) attr.Type {
	return types.Int64Type
}
