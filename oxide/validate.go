// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"

	"github.com/google/uuid"
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

// AddError adds an error to the validator's error collection.
func (v *Validator) AddError(err error) {
	v.err = errors.Join(v.err, err)
}

// MatchesPattern checks that a string matches the given regex pattern.
func (v *Validator) MatchesPattern(value, pattern, name string) bool {
	if value == "" {
		return true // empty values are handled by required checks
	}
	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		v.err = errors.Join(v.err, fmt.Errorf("invalid pattern for %s: %w", name, err))
		return false
	}
	if !matched {
		v.err = errors.Join(v.err, fmt.Errorf("%s must match pattern %s", name, pattern))
		return false
	}
	return true
}

// ValidEnum checks that a value is one of the allowed enum values.
func ValidEnum[T comparable](v *Validator, value T, allowed []T, name string) bool {
	for _, a := range allowed {
		if value == a {
			return true
		}
	}
	v.err = errors.Join(v.err, fmt.Errorf("%s has invalid value %v", name, value))
	return false
}

// ValidFormat checks that a string matches the expected format.
func (v *Validator) ValidFormat(value, format, name string) bool {
	if value == "" {
		return true // empty values are handled by required checks
	}

	var err error
	switch format {
	case "uuid":
		_, err = uuid.Parse(value)
		if err != nil {
			v.err = errors.Join(v.err, fmt.Errorf("%s must be a valid UUID: %w", name, err))
			return false
		}
	case "email":
		_, err = mail.ParseAddress(value)
		if err != nil {
			v.err = errors.Join(v.err, fmt.Errorf("%s must be a valid email address: %w", name, err))
			return false
		}
	case "uri", "url":
		_, err = url.ParseRequestURI(value)
		if err != nil {
			v.err = errors.Join(v.err, fmt.Errorf("%s must be a valid URI: %w", name, err))
			return false
		}
	case "ipv4":
		ip := net.ParseIP(value)
		if ip == nil || ip.To4() == nil {
			v.err = errors.Join(v.err, fmt.Errorf("%s must be a valid IPv4 address", name))
			return false
		}
	case "ipv6":
		ip := net.ParseIP(value)
		if ip == nil || ip.To4() != nil {
			v.err = errors.Join(v.err, fmt.Errorf("%s must be a valid IPv6 address", name))
			return false
		}
	case "ip":
		ip := net.ParseIP(value)
		if ip == nil {
			v.err = errors.Join(v.err, fmt.Errorf("%s must be a valid IP address", name))
			return false
		}
	case "hostname":
		if !isValidHostname(value) {
			v.err = errors.Join(v.err, fmt.Errorf("%s must be a valid hostname", name))
			return false
		}
	// date-time is handled by Go's time.Time type, so no runtime validation needed
	case "date-time", "date", "time":
		return true
	default:
		// Unknown format - skip validation
		return true
	}
	return true
}

// isValidHostname checks if a string is a valid hostname per RFC 1123.
func isValidHostname(s string) bool {
	if len(s) == 0 || len(s) > 253 {
		return false
	}
	// Hostname regex per RFC 1123
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	return hostnameRegex.MatchString(s)
}
