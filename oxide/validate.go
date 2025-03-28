// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"errors"
	"fmt"
	"reflect"
)

// Validator is a helper to validate the Client methods
type Validator struct {
	err error
}

// HasRequiredStr checks for an empty string
func (v *Validator) HasRequiredStr(value, name string) bool {
	if value == "" {
		v.err = errors.Join(v.err, fmt.Errorf("required value for %s is an empty string", name))
		return false
	}
	return true
}

// HasRequiredNum checks that a value is not nil
func (v *Validator) HasRequiredNum(value *int, name string) bool {
	if value == nil {
		v.err = errors.Join(v.err, fmt.Errorf("required value for %s is nil", name))
		return false
	}
	return true
}

// HasRequiredObj checks for a nil value.
// The argument must be a chan, func, interface, map,
// pointer, or slice value
func (v *Validator) HasRequiredObj(value any, name string) bool {
	// Unfortunately generics are a little tricky when
	// dealing with nil values, so we have to use reflect here.
	if value == nil || reflect.ValueOf(value).IsNil() {
		v.err = errors.Join(v.err, fmt.Errorf("required value for %s is nil", name))
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
