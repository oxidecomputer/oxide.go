// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"fmt"
	"reflect"
)

// Validator is a helper to validate the Client methods
type Validator struct {
	// TODO: Capture multiple errors
	err error
}

// HasRequiredStr checks for an empty string
func (v *Validator) HasRequiredStr(value string) bool {
	if value == "" {
		v.err = fmt.Errorf("required value is an empty string")
		return false
	}
	return true
}

// HasRequiredNum checks for an empty string
func (v *Validator) HasRequiredNum(value int) bool {
	if value == 0 {
		v.err = fmt.Errorf("required value is zero")
		return false
	}
	return true
}

// HasRequiredObj checks for a nil value
// The argument must be a chan, func, interface, map,
// pointer, or slice value
func (v *Validator) HasRequiredObj(value any) bool {
	// Unfortunately generics are a little tricky when
	// dealing with nil values, so we have to use reflect here.
	if value == nil || reflect.ValueOf(value).IsNil() {
		v.err = fmt.Errorf("required value is nil")
		return false
	}
	return true
}

// IsValid returns false if the Validator contains
// any validation errors
func (v *Validator) IsValid() bool {
	return v.err == nil
}

// Error is the string representation of a validation
// error
func (v *Validator) Error() string {
	return v.err.Error()
}
